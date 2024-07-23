package session

import (
	"context"
	"errors"
	"log"
	"math"
	"time"

	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/param"
	dbUtil "github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lnm/internal/ito"
	metrics "github.com/satimoto/go-lnm/internal/metric"
	"github.com/satimoto/go-lnm/internal/user"
	"github.com/satimoto/go-lnm/pkg/util"
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
		metrics.RecordError("LNM137", "Error in session", errors.New("authorizationID is nil"))
		log.Printf("LNM137: SessionUid=%v", session.Uid)

		// Get the last token authorization for the session token
		tokenAuthorization, err := r.TokenAuthorizationRepository.GetLastTokenAuthorizationByTokenID(ctx, session.TokenID)

		if err != nil {
			// No last token authorization found
			metrics.RecordError("LNM136", "Error last token authorization not found", err)
			log.Printf("LNM136: SessionUid=%v, TokenID=%v", session.Uid, session.TokenID)
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
		metrics.RecordError("LNM037", "Error retrieving user from session", err)
		log.Printf("LNM037: SessionUid=%v, UserID=%v", session.Uid, session.UserID)
		r.StopSession(ctx, session)
		return
	}

	connector, err := r.LocationRepository.GetConnector(ctx, session.ConnectorID)

	if err != nil {
		metrics.RecordError("LNM001", "Error retrieving session connector", err)
		log.Printf("LNM001: SessionUid=%v, ConnectorID=%v", session.Uid, session.ConnectorID)
		r.StopSession(ctx, session)
		return
	}

	tokenAuthorization, err := r.TokenAuthorizationRepository.GetTokenAuthorizationByAuthorizationID(ctx, session.AuthorizationID.String)

	if err != nil {
		metrics.RecordError("LNM127", "Error retrieving token authorization", err)
		log.Printf("LNM127: SessionUid=%v, AuthorizationID=%v", session.Uid, session.AuthorizationID.String)
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
			metrics.RecordError("LNM002", "Error retrieving session tariff", err)
			log.Printf("LNM002: SessionUid=%v, TariffID=%v", session.Uid, connector.TariffID.String)
			return
		}

		tariffIto := r.TariffResolver.CreateTariffIto(ctx, tariff)
		location, err := r.LocationRepository.GetLocation(ctx, session.LocationID)

		if err != nil {
			metrics.RecordError("LNM038", "Error retrieving session location", err)
			log.Printf("LNM038: SessionUid=%v, LocationID=%v", session.Uid, session.LocationID)
			return
		}

		timeLocation, err := time.LoadLocation(location.TimeZone.String)

		if err != nil {
			metrics.RecordError("LNM005", "Error loading time location", err)
			log.Printf("LNM005: TimeZone=%v", location.TimeZone.String)
			timeLocation, err = time.LoadLocation("UTC")
		}

		taxPercent := r.AccountResolver.GetTaxPercentByCountry(ctx, location.Country, dbUtil.GetEnvFloat64("DEFAULT_TAX_PERCENT", 19))
		invoiceInterval := calculateInvoiceInterval(connector.Wattage)
		log.Printf("Monitor session for %s, running every %f seconds", session.Uid, invoiceInterval.Seconds())

	invoiceLoop:
		for {
			// Wait for invoice interval
			time.Sleep(invoiceInterval)

			// Get latest session
			session, err = r.Repository.GetSession(ctx, session.ID)

			if err != nil {
				metrics.RecordError("LNM032", "Error retrieving session", err)
				log.Printf("LNM032: SessionUid=%v", session.Uid)
				continue
			}

			switch session.Status {
			case db.SessionStatusTypeCOMPLETED, db.SessionStatusTypeENDING, db.SessionStatusTypeINVALID, db.SessionStatusTypeINVOICED:
				// End monitoring, let the CDR issue the final invoice
				log.Printf("Ending session monitoring for %s", session.Uid)
				r.SendSessionUpdateNotification(user, session)
				break invoiceLoop
			case db.SessionStatusTypeACTIVE:
				// Session is active, calculate new invoice
				if ok := r.processInvoicePeriod(ctx, user, session, timeLocation, tariffIto, connector, taxPercent); !ok {
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

	if token, err := r.TokenRepository.GetToken(ctx, session.TokenID); err == nil && token.Type == db.TokenTypeOTHER {
		return r.OcpiService.StopSession(ctx, &ocpirpc.StopSessionRequest{
			AuthorizationId: session.AuthorizationID.String,
		})
	}

	return nil, errors.New("cannot remotely stop this session")
}

func (r *SessionResolver) processInvoicePeriod(ctx context.Context, sessionUser db.User, session db.Session, timeLocation *time.Location, tariffIto *ito.TariffIto, connector db.Connector, taxPercent float64) bool {
	sessionInvoices, err := r.Repository.ListSessionInvoicesBySessionID(ctx, session.ID)

	if err != nil {
		metrics.RecordError("LNM033", "Error retrieving session invoices", err)
		log.Printf("LNM033: SessionUid=%v", session.Uid)
		return true
	}

	/* TODO: Update how we handle unsettled invoices.
	//       Mobile devices are not reliable enought to pay invoices as they arrive
	//       to the device because it may be in a hibernated state with no network access.
	//       The session invoice is sent to the device and it should be paid once the
	//       device is woken for a background task or the app is opened.
	if hasUnsettledInvoices(sessionInvoices) {
		// Lock user tokens until all session invoices are settled
		err = r.UserResolver.RestrictUser(ctx, user)

		if err != nil {
			metrics.RecordError("LNM042", "Error restricting user", err)
			log.Printf("LNM042: SessionUID=%v, UserID=%v", session.Uid, session.UserID)
		}

		if _, err = r.StopSession(ctx, session); err == nil {
			// End invoice loop, let the cdr settle the session
			log.Printf("Session %s has unsettled invoices, stopping the session", session.Uid)

			return false
		}
	}*/

	timeNow := time.Now().UTC()
	delta := timeNow.Sub(session.LastUpdated).Minutes()
	log.Printf("Processing session %v with currency %v", session.Uid, tariffIto.Currency)
	log.Printf("%v: Kwh=%v, LastUpdated=%v, DeltaMinutes=%v", session.Uid, session.Kwh, session.LastUpdated.Format(time.RFC3339), delta)

	if session.TotalCost.Valid {
		log.Printf("%v: TotalCost=%v", session.Uid, session.TotalCost.Float64)
	}

	estimatedChargePower := user.GetEstimatedChargePower(sessionUser, connector)
	invoicedPriceFiat, _ := CalculatePriceInvoiced(sessionInvoices)
	sessionIto := r.CreateSessionIto(ctx, session)
	estimatedFiat, estimatedEnergy, estimatedTime := r.ProcessChargingPeriods(sessionIto, tariffIto, estimatedChargePower, timeLocation, timeNow)

	// Sanity check the estimated energy
	// If the estimated energy is over 50 kWh then flag the session and stop issuing invoices
	if estimatedEnergy >= 50 {
		log.Printf("Flagging session %v because of estimated energy", session.Uid)
		r.FlagSession(ctx, session)
		return false
	}

	// Sanity check the estimated time
	// If the estimated time is over 6 hours then flag the session and stop issuing invoices
	if estimatedTime >= 6 {
		log.Printf("Flagging session %v because of estimated time", session.Uid)
		r.FlagSession(ctx, session)
		return false
	}

	if estimatedFiat > invoicedPriceFiat {
		priceFiat := estimatedFiat - invoicedPriceFiat
		totalFiat, commissionFiat, taxFiat := CalculateCommission(estimatedFiat-invoicedPriceFiat, sessionUser.CommissionPercent, taxPercent)
		meteredTime := session.LastUpdated.Sub(sessionIto.StartDatetime).Hours()

		invoiceParams := util.InvoiceParams{
			Currency:       tariffIto.Currency,
			PriceFiat:      dbUtil.SqlNullFloat64(priceFiat),
			CommissionFiat: dbUtil.SqlNullFloat64(commissionFiat),
			TaxFiat:        dbUtil.SqlNullFloat64(taxFiat),
			TotalFiat:      dbUtil.SqlNullFloat64(totalFiat),
		}

		chargeParams := util.ChargeParams{
			EstimatedEnergy: math.Max(0, estimatedEnergy),
			EstimatedTime:   math.Max(0, estimatedTime),
			MeteredEnergy:   math.Max(0, sessionIto.TotalEnergy),
			MeteredTime:     math.Max(0, meteredTime),
		}

		r.IssueSessionInvoice(ctx, sessionUser, session, invoiceParams, chargeParams)
	}

	return true
}
