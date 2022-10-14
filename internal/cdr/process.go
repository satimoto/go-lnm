package cdr

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"time"

	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/param"
	dbUtil "github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lsp/internal/session"
	"github.com/satimoto/go-lsp/pkg/util"
)

func (r *CdrResolver) ProcessCdr(cdr db.Cdr) error {
	/** Cdr has been created.
	 *  Calculate final invoiced amount.
	 *  Issue final invoice or rebate if overpaid
	 */

	ctx := context.Background()
	authorizationId := cdr.AuthorizationID

	if !authorizationId.Valid {
		// There is no AuthorizationID set,
		// we cannot reconcile the session without it.
		// Flag the cdr to be looked at later.
		log.Printf("LSP035: Cdr AuthorizationID is nil")
		log.Printf("LSP035: CdrUid=%v", cdr.Uid)

		r.Repository.UpdateCdrIsFlaggedByUid(ctx, db.UpdateCdrIsFlaggedByUidParams{
			Uid:       cdr.Uid,
			IsFlagged: true,
		})

		sessions, err := r.SessionResolver.Repository.ListInProgressSessionsByUserID(ctx, cdr.UserID)

		if err != nil {
			dbUtil.LogOnError("LSP139", "Error retrieving in progress sessions", err)
			log.Printf("LSP139: UserID=%v", cdr.UserID)
			return errors.New("error retrieving in progress sessions")
		}

		// TODO: Should we close out all sessions or pick the best fitting session match
		if len(sessions) == 1 {
			// There is one in progress session, we can assume this is cdr session
			sess := sessions[0]

			// Check the session and cdr location/evs/connector matches
			if sess.AuthID == cdr.AuthID && sess.LocationID == cdr.LocationID && sess.EvseID == cdr.EvseID && sess.ConnectorID == cdr.ConnectorID {
				log.Printf("LSP035: Using matched session %v with authorization %v instead", sess.Uid, sess.AuthorizationID.String)
				authorizationId = sess.AuthorizationID
			}
		}

		if !authorizationId.Valid {
			for _, sess := range sessions {
				log.Printf("LSP035: Stopping session %v", sess.Uid)
				sessionParams := param.NewUpdateSessionByUidParams(sess)
				sessionParams.Status = db.SessionStatusTypeCOMPLETED
			
				_, err = r.SessionResolver.Repository.UpdateSessionByUid(ctx, sessionParams)

				r.SessionResolver.StopSession(ctx, sess)
			}
	
			return errors.New("cdr AuthorizationID is nil")	
		}
	}

	// TODO: How to deal with Sessions and CDRs with no AuthorizationID
	sess, err := r.SessionResolver.Repository.GetSessionByAuthorizationID(ctx, authorizationId.String)

	if err != nil {
		dbUtil.LogOnError("LSP043", "Error retrieving cdr session", err)
		log.Printf("LSP043: CdrUid=%v, AuthorizationID=%v", cdr.Uid, authorizationId)
		return errors.New("error retrieving cdr session")
	}

	user, err := r.SessionResolver.UserResolver.Repository.GetUser(ctx, sess.UserID)

	if err != nil {
		dbUtil.LogOnError("LSP044", "Error retrieving session user", err)
		log.Printf("LSP044: SessionUid=%v, UserID=%v", sess.Uid, sess.UserID)
		return errors.New("error retrieving session user")
	}

	location, err := r.SessionResolver.LocationRepository.GetLocation(ctx, sess.LocationID)

	if err != nil {
		dbUtil.LogOnError("LSP045", "Error retrieving session location", err)
		log.Printf("LSP045: SessionUid=%v, LocationID=%v", sess.Uid, sess.LocationID)
		return errors.New("error retrieving session location")
	}

	sessionInvoices, err := r.SessionResolver.Repository.ListSessionInvoices(ctx, sess.ID)

	if err != nil {
		dbUtil.LogOnError("LSP046", "Error retrieving session invoices", err)
		log.Printf("LSP046: SessionUid=%v", sess.Uid)
		return errors.New("error retrieving session invoices")
	}

	priceFiat, _ := session.CalculatePriceInvoiced(sessionInvoices)
	_, totalMsat := session.CalculateTotalInvoiced(sessionInvoices)

	taxPercent := r.SessionResolver.CountryAccountResolver.GetTaxPercentByCountry(ctx, location.Country, dbUtil.GetEnvFloat64("DEFAULT_TAX_PERCENT", 19))
	cdrTotalFiat := cdr.TotalCost

	switch {
	case cdrTotalFiat > priceFiat:
		// Issue final invoice
		tokenAuthorization, err := r.SessionResolver.TokenAuthorizationRepository.GetTokenAuthorizationByAuthorizationID(ctx, sess.AuthorizationID.String)

		if err != nil {
			dbUtil.LogOnError("LSP128", "Error retrieving token authorization", err)
			log.Printf("LSP128: SessionUid=%v, AuthorizationID=%v", sess.Uid, sess.AuthorizationID.String)
			return errors.New("error retrieving token authorization")
		}

		invoicePriceFiat := cdrTotalFiat - priceFiat
		invoiceTotalFiat, invoiceCommissionFiat, invoiceTaxFiat := session.CalculateCommission(cdrTotalFiat-priceFiat, user.CommissionPercent, taxPercent)
		sessionInvoice := r.SessionResolver.IssueSessionInvoice(ctx, user, sess, tokenAuthorization, util.InvoiceParams{
			PriceFiat:      dbUtil.SqlNullFloat64(invoicePriceFiat),
			CommissionFiat: dbUtil.SqlNullFloat64(invoiceCommissionFiat),
			TaxFiat:        dbUtil.SqlNullFloat64(invoiceTaxFiat),
			TotalFiat:      dbUtil.SqlNullFloat64(invoiceTotalFiat),
		})

		if sessionInvoice != nil {
			sessionInvoices = append(sessionInvoices, *sessionInvoice)
			_, totalMsat = session.CalculateTotalInvoiced(sessionInvoices)
		}
	case cdrTotalFiat <= priceFiat:
		if cdrTotalFiat < priceFiat {
			// Issue rebate if overpaid
			// TODO: This should be launched as a goroutine to force completion/retries
			rebateTotalFiat := priceFiat - cdrTotalFiat
			rebatePriceFiat, rebateCommissionFiat, rebateTaxFiat := session.ReverseCommission(rebateTotalFiat, user.CommissionPercent, taxPercent)

			invoiceParams := util.InvoiceParams{
				PriceFiat:      dbUtil.SqlNullFloat64(rebatePriceFiat),
				CommissionFiat: dbUtil.SqlNullFloat64(rebateCommissionFiat),
				TaxFiat:        dbUtil.SqlNullFloat64(rebateTaxFiat),
				TotalFiat:      dbUtil.SqlNullFloat64(rebateTotalFiat),
			}

			if invoiceRequest, err := r.IssueInvoiceRequest(ctx, user.ID, "REBATE", sess.Uid, sess.Currency, invoiceParams); err == nil {
				updateSessionByUidParams := param.NewUpdateSessionByUidParams(sess)
				updateSessionByUidParams.InvoiceRequestID = dbUtil.SqlNullInt64(invoiceRequest.ID)

				_, err := r.SessionResolver.Repository.UpdateSessionByUid(ctx, updateSessionByUidParams)

				if err != nil {
					dbUtil.LogOnError("LSP117", "Error updating session", err)
					log.Printf("LSP117: Params=%v", updateSessionByUidParams)
					return errors.New("error updating session")
				}
			}
		}

		r.SessionResolver.SendSessionUpdateNotification(user, sess)
	}

	// Set session as invoiced
	sessionParams := param.NewUpdateSessionByUidParams(sess)
	sessionParams.Status = db.SessionStatusTypeINVOICED

	_, err = r.SessionResolver.Repository.UpdateSessionByUid(ctx, sessionParams)

	if err != nil {
		dbUtil.LogOnError("LSP133", "Error updating session", err)
		log.Printf("LSP133: Params=%#v", sessionParams)
	}

	// Issue invoice request to circuit user
	circuitPercent := dbUtil.GetEnvFloat64("CIRCUIT_PERCENT", 0.5)
	circuitAmountMsat := int64((float64(totalMsat) / 100.0) * circuitPercent)

	if user.CircuitUserID.Valid && circuitAmountMsat > 0 {
		// TODO: This should be launched as a goroutine to force completion/retries
		releaseDate := time.Now().Add(time.Duration(rand.Intn(120)) * time.Minute)
		invoiceParams := util.InvoiceParams{
			TotalMsat:   dbUtil.SqlNullInt64(circuitAmountMsat),
			ReleaseDate: dbUtil.SqlNullTime(releaseDate),
		}

		_, err := r.IssueInvoiceRequest(ctx, user.CircuitUserID.Int64, "CIRCUIT", "Satsback", sess.Currency, invoiceParams)

		return err
	}

	return nil
}
