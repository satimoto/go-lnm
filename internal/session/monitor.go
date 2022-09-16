package session

import (
	"context"
	"log"
	"time"

	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/internal/tariff"
	"github.com/satimoto/go-ocpi/ocpirpc"
)

func (r *SessionResolver) StartSessionMonitor(ctx context.Context, session db.Session) {
	/** Session has been created.
	 *  Send SessionUpdate notification to user.
	 *  Calculate invoiced amount.
	 *  Define invoice period based on connector wattage
	 *  Periodically calculate session total, issue an invoice to the user.
	 *  Monitor issued invoices, if invoices go unpaid, cancel session.
	 */

	user, err := r.UserResolver.Repository.GetUser(ctx, session.UserID)

	if err != nil {
		util.LogOnError("LSP037", "Error retrieving user from session", err)
		log.Printf("LSP037: SessionUid=%v, UserID=%v", session.Uid, session.UserID)
		return
	}

	r.SendSessionUpdateNotification(user, session)

	connector, err := r.LocationRepository.GetConnector(ctx, session.ConnectorID)

	if err != nil {
		util.LogOnError("LSP001", "Error retrieving session connector", err)
		log.Printf("LSP001: SessionUid=%v, ConnectorID=%v", session.Uid, session.ConnectorID)
		return
	}

	if connector.TariffID.Valid {
		tariff, err := r.TariffResolver.Repository.GetTariffByUid(ctx, connector.TariffID.String)

		if err != nil {
			util.LogOnError("LSP002", "Error retrieving session tariff", err)
			log.Printf("LSP002: SessionUid=%v, TariffID=%v", session.Uid, connector.TariffID.String)
			return
		}

		tariffIto := r.TariffResolver.CreateTariffIto(ctx, tariff)
		location, err := r.LocationRepository.GetLocation(ctx, session.LocationID)

		if err != nil {
			util.LogOnError("LSP038", "Error retrieving session location", err)
			log.Printf("LSP038: SessionUid=%v, LocationID=%v", session.Uid, session.LocationID)
			return
		}

		taxPercent := r.CountryAccountResolver.GetTaxPercentByCountry(ctx, location.Country, util.GetEnvFloat64("DEFAULT_TAX_PERCENT", 19))
		invoiceInterval := calculateInvoiceInterval(connector.Wattage)
		log.Printf("Monitor session for %s, running every %v seconds", session.Uid, invoiceInterval/time.Second)

	invoiceLoop:
		for {
			// Wait for invoice interval
			time.Sleep(invoiceInterval)

			// Get latest session
			session, err = r.Repository.GetSessionByUid(ctx, session.Uid)

			if err != nil {
				util.LogOnError("LSP032", "Error retrieving session", err)
				log.Printf("LSP032: SessionUid=%v", session.Uid)
				continue
			}

			switch session.Status {
			case db.SessionStatusTypeCOMPLETED, db.SessionStatusTypeINVALID:
				// End monitoring, let the CDR issue the final invoice
				log.Printf("Ending session monitoring for %s", session.Uid)
				break invoiceLoop
			case db.SessionStatusTypeACTIVE:
				// Session is active, calculate new invoice
				if ok := r.processInvoicePeriod(ctx, user, session, tariffIto, taxPercent); !ok {
					log.Printf("Ending session monitoring for %s with errors", session.Uid)
					break invoiceLoop
				}
			}
		}
	}
}

func (r *SessionResolver) processInvoicePeriod(ctx context.Context, user db.User, session db.Session, tariffIto *tariff.TariffIto, taxPercent float64) bool {
	sessionInvoices, err := r.Repository.ListSessionInvoices(ctx, session.ID)

	if err != nil {
		util.LogOnError("LSP033", "Error retrieving session invoices", err)
		log.Printf("LSP033: SessionUid=%v", session.Uid)
		return true
	}

	if hasUnsettledInvoices(sessionInvoices) {
		// Kill session
		// Suspend tokens until balance is settled
		// TODO: handle expired invoices, reissue invoices on request
		log.Printf("Session %s has unsettled invoices, stopping the session", session.Uid)

		r.OcpiService.StopSession(ctx, &ocpirpc.StopSessionRequest{
			SessionUid: session.Uid,
		})

		// Lock user tokens until all session invoices are settled
		err = r.UserResolver.RestrictUser(ctx, user)

		if err != nil {
			util.LogOnError("LSP042", "Error restricting user", err)
			log.Printf("LSP042: SessionUID=%v, UserID=%v", session.Uid, session.UserID)
		}

		// End invoice loop
		return false
	}

	amountFiat, _ := CalculateAmountInvoiced(sessionInvoices)
	sessionIto := r.CreateSessionIto(ctx, session)
	sessionAmount := r.ProcessChargingPeriods(sessionIto, tariffIto, time.Now().UTC())
	totalAmount, _, _ := CalculateCommission(sessionAmount, user.CommissionPercent, taxPercent)

	if totalAmount > amountFiat {
		invoiceAmount, invoiceCommission, invoiceTax := CalculateCommission(totalAmount-amountFiat, user.CommissionPercent, taxPercent)

		r.IssueSessionInvoice(ctx, user, session, invoiceAmount, invoiceCommission, invoiceTax)
	}

	return true
}
