package session

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/param"
	dbUtil "github.com/satimoto/go-datastore/pkg/util"
	metrics "github.com/satimoto/go-lsp/internal/metric"
	"github.com/satimoto/go-lsp/internal/tariff"
	"github.com/satimoto/go-lsp/pkg/util"
	"github.com/satimoto/go-ocpi/ocpirpc"
)

func (r *SessionResolver) StartSessionMonitor(session db.Session) {
	/** Session has been created.
	 *  Send SessionUpdate notification to user.
	 *  Calculate invoiced amount.
	 *  Define invoice period based on connector wattage
	 *  Periodically calculate session total, issue an invoice to the user.
	 *  Monitor issued invoices, if invoices go unpaid, cancel session.
	 */

	metricSessionMonitoringGoroutines.Inc()
	defer metricSessionMonitoringGoroutines.Dec()

	ctx := context.Background()

	if !session.AuthorizationID.Valid {
		// There is no AuthorizationID set, flag the session.
		metrics.RecordError("LSP137", "Error in session", errors.New("authorizationID is nil"))
		log.Printf("LSP137: SessionUid=%v", session.Uid)

		// Get the last token authorization for the session token
		tokenAuthorization, err := r.TokenAuthorizationRepository.GetLastTokenAuthorizationByTokenID(ctx, session.TokenID)

		if err != nil {
			// No last token authorization found
			metrics.RecordError("LSP136", "Error last token authorization not found", err)
			log.Printf("LSP136: SessionUid=%v, TokenID=%v", session.Uid, session.TokenID)
			r.StopSession(ctx, session)
			return
		}

		// Manually set the session authorizationID
		updateSessionByUidParams := param.NewUpdateSessionByUidParams(session)
		updateSessionByUidParams.AuthorizationID = dbUtil.SqlNullString(tokenAuthorization.AuthorizationID)
		updateSessionByUidParams.IsFlagged = true

		if updatedSession, err := r.Repository.UpdateSessionByUid(ctx, updateSessionByUidParams); err == nil {
			session = updatedSession
		}

		RecordFlaggedSession()
	}

	user, err := r.UserResolver.Repository.GetUser(ctx, session.UserID)

	if err != nil {
		metrics.RecordError("LSP037", "Error retrieving user from session", err)
		log.Printf("LSP037: SessionUid=%v, UserID=%v", session.Uid, session.UserID)
		r.StopSession(ctx, session)
		return
	}

	connector, err := r.LocationRepository.GetConnector(ctx, session.ConnectorID)

	if err != nil {
		metrics.RecordError("LSP001", "Error retrieving session connector", err)
		log.Printf("LSP001: SessionUid=%v, ConnectorID=%v", session.Uid, session.ConnectorID)
		r.StopSession(ctx, session)
		return
	}

	tokenAuthorization, err := r.TokenAuthorizationRepository.GetTokenAuthorizationByAuthorizationID(ctx, session.AuthorizationID.String)

	if err != nil {
		metrics.RecordError("LSP127", "Error retrieving token authorization", err)
		log.Printf("LSP127: SessionUid=%v, AuthorizationID=%v", session.Uid, session.AuthorizationID.String)
		r.StopSession(ctx, session)
		return
	}

	if !tokenAuthorization.Authorized {
		log.Printf("Ending unauthorized session %s", session.Uid)
		r.StopSession(ctx, session)
		return
	}

	r.SendSessionUpdateNotification(user, session)

	if connector.TariffID.Valid {
		tariff, err := r.TariffResolver.Repository.GetTariffByUid(ctx, connector.TariffID.String)

		if err != nil {
			metrics.RecordError("LSP002", "Error retrieving session tariff", err)
			log.Printf("LSP002: SessionUid=%v, TariffID=%v", session.Uid, connector.TariffID.String)
			return
		}

		tariffIto := r.TariffResolver.CreateTariffIto(ctx, tariff)
		location, err := r.LocationRepository.GetLocation(ctx, session.LocationID)

		if err != nil {
			metrics.RecordError("LSP038", "Error retrieving session location", err)
			log.Printf("LSP038: SessionUid=%v, LocationID=%v", session.Uid, session.LocationID)
			return
		}

		timeLocation, err := time.LoadLocation(location.TimeZone.String)

		if err != nil {
			metrics.RecordError("LSP005", "Error loading time location", err)
			log.Printf("LSP005: TimeZone=%v", location.TimeZone.String)
			timeLocation, err = time.LoadLocation("UTC")
		}

		taxPercent := r.CountryAccountResolver.GetTaxPercentByCountry(ctx, location.Country, dbUtil.GetEnvFloat64("DEFAULT_TAX_PERCENT", 19))
		invoiceInterval := calculateInvoiceInterval(connector.Wattage)
		log.Printf("Monitor session for %s, running every %f seconds", session.Uid, invoiceInterval.Seconds())

		if session.Status == db.SessionStatusTypePENDING {
			go r.waitForSessionTimeout(user, session.Uid, 90*time.Second)
		}

	invoiceLoop:
		for {
			// Wait for invoice interval
			time.Sleep(invoiceInterval)

			// Get latest session
			session, err = r.Repository.GetSessionByUid(ctx, session.Uid)

			if err != nil {
				metrics.RecordError("LSP032", "Error retrieving session", err)
				log.Printf("LSP032: SessionUid=%v", session.Uid)
				continue
			}

			switch session.Status {
			case db.SessionStatusTypeCOMPLETED, db.SessionStatusTypeINVALID, db.SessionStatusTypeINVOICED:
				// End monitoring, let the CDR issue the final invoice
				log.Printf("Ending session monitoring for %s", session.Uid)
				break invoiceLoop
			case db.SessionStatusTypeACTIVE:
				// Session is active, calculate new invoice
				if ok := r.processInvoicePeriod(ctx, user, session, tokenAuthorization, timeLocation, tariffIto, connector.Wattage, taxPercent); !ok {
					log.Printf("Ending session monitoring for %s with errors", session.Uid)
					break invoiceLoop
				}
			}
		}
	}
}

func (r *SessionResolver) FlagSession(ctx context.Context, session db.Session) {
	r.Repository.UpdateSessionIsFlaggedByUid(ctx, db.UpdateSessionIsFlaggedByUidParams{
		Uid:       session.Uid,
		IsFlagged: true,
	})

	RecordFlaggedSession()
}

func (r *SessionResolver) StopSession(ctx context.Context, session db.Session) (*ocpirpc.StopSessionResponse, error) {
	r.FlagSession(ctx, session)

	if session.AuthMethod == db.AuthMethodTypeAUTHREQUEST {
		return r.OcpiService.StopSession(ctx, &ocpirpc.StopSessionRequest{
			SessionUid: session.Uid,
		})
	}

	return nil, errors.New("cannot remotely stop this session")
}

func (r *SessionResolver) processInvoicePeriod(ctx context.Context, user db.User, session db.Session, tokenAuthorization db.TokenAuthorization, timeLocation *time.Location, tariffIto *tariff.TariffIto, connectorWattage int32, taxPercent float64) bool {
	sessionInvoices, err := r.Repository.ListSessionInvoices(ctx, session.ID)

	if err != nil {
		metrics.RecordError("LSP033", "Error retrieving session invoices", err)
		log.Printf("LSP033: SessionUid=%v", session.Uid)
		return true
	}

	if hasUnsettledInvoices(sessionInvoices) {
		// Lock user tokens until all session invoices are settled
		err = r.UserResolver.RestrictUser(ctx, user)

		if err != nil {
			metrics.RecordError("LSP042", "Error restricting user", err)
			log.Printf("LSP042: SessionUID=%v, UserID=%v", session.Uid, session.UserID)
		}

		if session.AuthMethod == db.AuthMethodTypeAUTHREQUEST {
			// Kill session
			// TODO: handle expired invoices, reissue invoices on request
			log.Printf("Session %s has unsettled invoices, stopping the session", session.Uid)
			r.StopSession(ctx, session)

			// End invoice loop, let the cdr settle the session
			return false
		}
	}

	timeNow := time.Now().UTC()
	delta := timeNow.Sub(session.LastUpdated).Minutes()
	log.Printf("Processing session %v with currency %v", session.Uid, session.Currency)
	log.Printf("%v: Kwh=%v, LastUpdated=%v, DeltaMinutes=%v", session.Uid, session.Kwh, session.LastUpdated.Format(time.RFC3339), delta)

	if session.TotalCost.Valid {
		log.Printf("%v: TotalCost=%v", session.Uid, session.TotalCost.Float64)
	}

	priceFiat, _ := CalculatePriceInvoiced(sessionInvoices)
	sessionIto := r.CreateSessionIto(ctx, session)
	sessionFiat := r.ProcessChargingPeriods(sessionIto, tariffIto, connectorWattage, timeLocation, timeNow)

	if sessionFiat > priceFiat {
		invoicePriceFiat := sessionFiat - priceFiat
		invoiceTotalFiat, invoiceCommissionFiat, invoiceTaxFiat := CalculateCommission(sessionFiat-priceFiat, user.CommissionPercent, taxPercent)

		r.IssueSessionInvoice(ctx, user, session, tokenAuthorization, util.InvoiceParams{
			PriceFiat:      dbUtil.SqlNullFloat64(invoicePriceFiat),
			CommissionFiat: dbUtil.SqlNullFloat64(invoiceCommissionFiat),
			TaxFiat:        dbUtil.SqlNullFloat64(invoiceTaxFiat),
			TotalFiat:      dbUtil.SqlNullFloat64(invoiceTotalFiat),
		})
	}

	return true
}

func (r *SessionResolver) waitForSessionTimeout(user db.User, sessionUid string, timeout time.Duration) {
	time.Sleep(timeout)

	ctx := context.Background()
	session, err := r.Repository.GetSessionByUid(ctx, sessionUid)

	if err != nil {
		metrics.RecordError("LSP155", "Error getting session", err)
		log.Printf("LSP155: SessionUid=%v", sessionUid)
		return
	}

	if session.Status == db.SessionStatusTypePENDING {
		updateSessionByUidParams := param.NewUpdateSessionByUidParams(session)
		updateSessionByUidParams.Status = db.SessionStatusTypeINVALID

		updatedSession, err := r.Repository.UpdateSessionByUid(ctx, updateSessionByUidParams)

		if err != nil {
			metrics.RecordError("LSP156", "Error updating session", err)
			log.Printf("LSP156: Params=%#v", updateSessionByUidParams)
			return
		}

		r.SendSessionUpdateNotification(user, updatedSession)
	}
}
