package session

import (
	"context"
	"log"
	"time"

	"github.com/satimoto/go-datastore/db"
	"github.com/satimoto/go-lsp/internal/util"
	"github.com/satimoto/go-ocpi-api/ocpirpc/commandrpc"
	"github.com/satimoto/go-ocpi-api/ocpirpc/tokenrpc"
)

func (r *SessionResolver) MonitorSession(ctx context.Context, session db.Session) {
	/** Session has been started.
	 *  Calculate invoiced amount.
	 *  Define invoice period based on connector wattage
	 *  Periodically calculate session total, issue an invoice to the user.
	 *  Monitor issued invoices, if invoices go unpaid, cancel session.
	 */

	if session.Status == db.SessionStatusTypeACTIVE || session.Status == db.SessionStatusTypePENDING {
		connector, err := r.LocationResolver.Repository.GetConnector(ctx, session.ConnectorID)

		if err != nil {
			util.LogOnError("LSP001", "Error retreiving session connector", err)
			log.Printf("LSP001: SessionUid=%v, ConnectorID=%v", session.Uid, session.ConnectorID)
			return
		}

		if connector.TariffID.Valid {
			tariff, err := r.TariffResolver.Repository.GetTariffByUid(ctx, connector.TariffID.String)

			if err != nil {
				util.LogOnError("LSP002", "Error retreiving session tariff", err)
				log.Printf("LSP002: SessionUid=%v, TariffID=%v", session.Uid, connector.TariffID.String)
				return
			}

			tariffIto := r.TariffResolver.CreateTariffIto(ctx, tariff)
			invoiceInterval := calculateInvoiceInterval(connector.Wattage)

			for {
				// Wait for invoice interval
				time.Sleep(invoiceInterval)

				// Get latest session
				session, err = r.Repository.GetSessionByUid(ctx, session.Uid)

				if err != nil {
					util.LogOnError("LSP032", "Error retreiving session", err)
					log.Printf("LSP032: SessionUid=%v", session.Uid)
					continue
				}
	
				if session.Status == db.SessionStatusTypeCOMPLETED || session.Status == db.SessionStatusTypeINVALID {
					// End monitoring, let the CDR issue the final invoice
					break
				}

				sessionInvoices, err := r.Repository.ListSessionInvoices(ctx, session.ID)

				if err != nil {
					util.LogOnError("LSP033", "Error retreiving session invoices", err)
					log.Printf("LSP033: SessionUid=%v", session.Uid)
					continue
				}

				if hasUnsettledInvoices(sessionInvoices) {
					// Kill session
					// Suspend tokens until balance is settled
					// TODO: handle expired invoices, reissue invoices on request
					r.OcpiService.GetCommandClient().StopSession(ctx, &commandrpc.StopSessionRequest{
						SessionUid: session.Uid,
					})

					user, err := r.UserResolver.Repository.GetUserByTokenID(ctx, session.TokenID)

					if err != nil {
						util.LogOnError("LSP035", "Error retreiving token user", err)
						log.Printf("LSP035: TokenID=%v", session.TokenID)
						break
					}	

					r.OcpiService.GetTokenClient().UpdateTokens(ctx, &tokenrpc.UpdateTokensRequest{
						UserId: user.ID,
						Allowed: string(db.TokenAllowedTypeNOCREDIT),
						Whitelist: string(db.TokenWhitelistTypeNEVER),
					})
					
					break
				}

				amountInvoiced := calculateAmountInvoiced(sessionInvoices)
				sessionIto := r.CreateSessionIto(ctx, session)
				latestAmount := r.ProcessChargingPeriods(sessionIto, tariffIto, time.Now())

				if latestAmount > amountInvoiced {
					amountFiat := amountInvoiced - latestAmount
					amountMsat := int64(amountFiat * 2500000)
					// TODO: get exchange rate

					r.IssueLightningInvoice(ctx, session, amountFiat, amountMsat)
				}
			}
		}
	}
}
