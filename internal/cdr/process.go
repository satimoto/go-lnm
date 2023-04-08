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
	metrics "github.com/satimoto/go-lnm/internal/metric"
	"github.com/satimoto/go-lnm/internal/session"
	"github.com/satimoto/go-lnm/internal/user"
	"github.com/satimoto/go-lnm/pkg/util"
)

func (r *CdrResolver) ProcessCdr(cdr db.Cdr) error {
	/** Cdr has been created.
	 *  Calculate final invoiced amount.
	 *  Issue final invoice or rebate if overpaid
	 */

	ctx := context.Background()
	authorizationId := cdr.AuthorizationID
	cdrIsFlagged := false

	if !authorizationId.Valid {
		// There is no AuthorizationID set,
		// we cannot reconcile the session without it.
		// Flag the cdr to be looked at later.
		log.Printf("LNM035: Cdr AuthorizationID is nil")
		log.Printf("LNM035: CdrUid=%v", cdr.Uid)

		cdrIsFlagged = true

		sessions, err := r.SessionResolver.Repository.ListInProgressSessionsByUserID(ctx, cdr.UserID)

		if err != nil {
			metrics.RecordError("LNM139", "Error retrieving in progress sessions", err)
			log.Printf("LNM139: UserID=%v", cdr.UserID)
			return errors.New("error retrieving in progress sessions")
		}

		// TODO: Should we close out all sessions or pick the best fitting session match
		if len(sessions) == 1 {
			// There is one in progress session, we can assume this is cdr session
			sess := sessions[0]

			// Check the session and cdr location/evs/connector matches
			if sess.AuthID == cdr.AuthID && sess.LocationID == cdr.LocationID && sess.EvseID == cdr.EvseID && sess.ConnectorID == cdr.ConnectorID {
				log.Printf("LNM035: Using matched session %v with authorization %v instead", sess.Uid, sess.AuthorizationID.String)
				authorizationId = sess.AuthorizationID
			}
		}

		if !authorizationId.Valid {
			for _, sess := range sessions {
				log.Printf("LNM035: Stopping session %v", sess.Uid)
				sessionParams := param.NewUpdateSessionByUidParams(sess)
				sessionParams.Status = db.SessionStatusTypeCOMPLETED

				_, err = r.SessionResolver.Repository.UpdateSessionByUid(ctx, sessionParams)

				r.SessionResolver.StopSession(ctx, sess)
			}

			return errors.New("cdr authorization ID is nil")
		}
	}

	tariffs, err := r.SessionResolver.TariffResolver.Repository.ListTariffsByCdr(ctx, dbUtil.SqlNullInt64(cdr.ID))

	if err != nil {
		metrics.RecordError("LNM157", "Error listing cdr tariffs", err)
		log.Printf("LNM157: CdrID=%v", cdr.ID)
	}

	// Flag the cdr if the cdr has a cost but no tariffs
	cdrIsFlagged = cdrIsFlagged || (cdr.TotalCost > 0 && len(tariffs) == 0)

	if cdrIsFlagged {
		r.Repository.UpdateCdrIsFlaggedByUid(ctx, db.UpdateCdrIsFlaggedByUidParams{
			Uid:       cdr.Uid,
			IsFlagged: cdrIsFlagged,
		})

		metricCdrsFlaggedTotal.Inc()
	}

	// TODO: How to deal with Sessions and CDRs with no AuthorizationID
	sess, err := r.SessionResolver.Repository.GetSessionByAuthorizationID(ctx, authorizationId.String)

	if err != nil {
		metrics.RecordError("LNM043", "Error retrieving cdr session", err)
		log.Printf("LNM043: CdrUid=%v, AuthorizationID=%v", cdr.Uid, authorizationId)
		return errors.New("error retrieving cdr session")
	}

	sessionUser, err := r.SessionResolver.UserResolver.Repository.GetUser(ctx, sess.UserID)

	if err != nil {
		metrics.RecordError("LNM044", "Error retrieving session user", err)
		log.Printf("LNM044: SessionUid=%v, UserID=%v", sess.Uid, sess.UserID)
		return errors.New("error retrieving session user")
	}

	location, err := r.SessionResolver.LocationRepository.GetLocation(ctx, sess.LocationID)

	if err != nil {
		metrics.RecordError("LNM045", "Error retrieving session location", err)
		log.Printf("LNM045: SessionUid=%v, LocationID=%v", sess.Uid, sess.LocationID)
		return errors.New("error retrieving session location")
	}

	sessionInvoices, err := r.SessionResolver.Repository.ListSessionInvoicesBySessionID(ctx, sess.ID)

	if err != nil {
		metrics.RecordError("LNM046", "Error retrieving session invoices", err)
		log.Printf("LNM046: SessionUid=%v", sess.Uid)
		return errors.New("error retrieving session invoices")
	}

	priceFiat, _ := session.CalculatePriceInvoiced(sessionInvoices)

	taxPercent := r.SessionResolver.AccountResolver.GetTaxPercentByCountry(ctx, location.Country, dbUtil.GetEnvFloat64("DEFAULT_TAX_PERCENT", 19))
	cdrTotalFiat := cdr.TotalCost
	cdrTotalEnergy := cdr.TotalEnergy
	cdrTotalTime := cdr.TotalTime

	// The cdr TotalCost might be 0. If so, we should check the TotalEnergy, TotalTime and TotalParkingTime
	if cdrTotalFiat == 0 && len(tariffs) > 0 {
		tariffIto := r.SessionResolver.TariffResolver.CreateTariffIto(ctx, tariffs[0])
		sessionIto := r.CreateSessionIto(ctx, cdr)

		connector, err := r.SessionResolver.LocationRepository.GetConnector(ctx, sess.ConnectorID)

		if err != nil {
			metrics.RecordError("LNM158", "Error getting session connector", err)
			log.Printf("LNM158: SessionUid=%v, ConnectorID=%v", sess.Uid, sess.ConnectorID)
			return errors.New("error gettings session connector")
		}

		timeLocation, err := time.LoadLocation(location.TimeZone.String)

		if err != nil {
			metrics.RecordError("LNM159", "Error loading time location", err)
			log.Printf("LNM159: TimeZone=%v", location.TimeZone.String)
			timeLocation, _ = time.LoadLocation("UTC")
		}

		estimatedChargePower := user.GetEstimatedChargePower(sessionUser, connector)

		cdrTotalFiat, cdrTotalEnergy, cdrTotalTime = r.SessionResolver.ProcessChargingPeriods(sessionIto, tariffIto, estimatedChargePower, timeLocation, cdr.LastUpdated)
	}

	// Set session as invoiced
	sessionParams := param.NewUpdateSessionByUidParams(sess)
	sessionParams.Status = db.SessionStatusTypeINVOICED

	if updatedSession, err := r.SessionResolver.Repository.UpdateSessionByUid(ctx, sessionParams); err == nil {
		sess = updatedSession
	}

	if cdrTotalFiat > 0 {
		chargeParams := util.ChargeParams{
			EstimatedEnergy: cdrTotalEnergy,
			EstimatedTime:   cdrTotalTime,
			MeteredEnergy:   cdrTotalEnergy,
			MeteredTime:     cdrTotalTime,
		}

		if cdrTotalFiat > priceFiat {
			// Issue final invoice
			invoicePriceFiat := cdrTotalFiat - priceFiat
			invoiceTotalFiat, invoiceCommissionFiat, invoiceTaxFiat := session.CalculateCommission(invoicePriceFiat, sessionUser.CommissionPercent, taxPercent)

			invoiceParams := util.InvoiceParams{
				PriceFiat:      dbUtil.SqlNullFloat64(invoicePriceFiat),
				CommissionFiat: dbUtil.SqlNullFloat64(invoiceCommissionFiat),
				TaxFiat:        dbUtil.SqlNullFloat64(invoiceTaxFiat),
				TotalFiat:      dbUtil.SqlNullFloat64(invoiceTotalFiat),
			}

			r.SessionResolver.IssueSessionInvoice(ctx, sessionUser, sess, invoiceParams, chargeParams)
		} else if cdrTotalFiat <= priceFiat {
			if cdrTotalFiat < priceFiat {
				// Issue rebate if overpaid
				// TODO: This should be launched as a goroutine to force completion/retries
				rebateTotalFiat := priceFiat - cdrTotalFiat
				rebatePriceFiat, rebateCommissionFiat, rebateTaxFiat := session.ReverseCommission(rebateTotalFiat, sessionUser.CommissionPercent, taxPercent)

				invoiceParams := util.InvoiceParams{
					PriceFiat:      dbUtil.SqlNullFloat64(rebatePriceFiat),
					CommissionFiat: dbUtil.SqlNullFloat64(rebateCommissionFiat),
					TaxFiat:        dbUtil.SqlNullFloat64(rebateTaxFiat),
					TotalFiat:      dbUtil.SqlNullFloat64(rebateTotalFiat),
				}

				r.IssueRebate(ctx, sess, sessionUser.ID, invoiceParams, chargeParams)
			}

			r.SessionResolver.SendSessionUpdateNotification(sessionUser, sess)
		}

		// Issue invoice request for session confirmation
		if sess.IsConfirmed {
			confirmationPercent := 0.1
			confirmationTotalFiat := (cdrTotalFiat / 100.0) * confirmationPercent
			confirmationPriceFiat, confirmationCommissionFiat, confirmationTaxFiat := session.ReverseCommission(confirmationTotalFiat, sessionUser.CommissionPercent, taxPercent)

			invoiceParams := util.InvoiceParams{
				PriceFiat:      dbUtil.SqlNullFloat64(confirmationPriceFiat),
				CommissionFiat: dbUtil.SqlNullFloat64(confirmationCommissionFiat),
				TaxFiat:        dbUtil.SqlNullFloat64(confirmationTaxFiat),
				TotalFiat:      dbUtil.SqlNullFloat64(confirmationTotalFiat),
			}

			r.IssueInvoiceRequest(ctx, sessionUser.ID, &sess.ID, "SESSION_CONFIRMED", sess.Currency, "Satimoto: Confirmed", invoiceParams)
		}

		// Issue invoice request based on location charge count
		if cdrsCount, err := r.Repository.CountCdrsByLocationID(ctx, cdr.LocationID); err == nil && cdrsCount == 1 {
			totalMsat := int64(21000)
			confirmationPriceMsat, confirmationCommissionMsat, confirmationTaxMsat := session.ReverseCommissionInt64(totalMsat, sessionUser.CommissionPercent, taxPercent)

			invoiceParams := util.InvoiceParams{
				PriceMsat:      dbUtil.SqlNullInt64(confirmationPriceMsat),
				CommissionMsat: dbUtil.SqlNullInt64(confirmationCommissionMsat),
				TaxMsat:        dbUtil.SqlNullInt64(confirmationTaxMsat),
				TotalMsat:      dbUtil.SqlNullInt64(totalMsat),
			}

			r.IssueInvoiceRequest(ctx, sessionUser.ID, &sess.ID, "FIRST_LOCATION_CHARGE", sess.Currency, "Satimoto: First", invoiceParams)
		}

		// Issue invoice request based on user charge count
		if cdrsCount, err := r.Repository.CountCdrsByUserID(ctx, cdr.UserID); err == nil {
			if cdrsCount == 1 {
				totalMsat := int64(21000)
				confirmationPriceMsat, confirmationCommissionMsat, confirmationTaxMsat := session.ReverseCommissionInt64(totalMsat, sessionUser.CommissionPercent, taxPercent)

				invoiceParams := util.InvoiceParams{
					PriceMsat:      dbUtil.SqlNullInt64(confirmationPriceMsat),
					CommissionMsat: dbUtil.SqlNullInt64(confirmationCommissionMsat),
					TaxMsat:        dbUtil.SqlNullInt64(confirmationTaxMsat),
					TotalMsat:      dbUtil.SqlNullInt64(totalMsat),
				}

				r.IssueInvoiceRequest(ctx, sessionUser.ID, &sess.ID, "FIRST_USER_CHARGE", sess.Currency, "Satimoto: Hello", invoiceParams)
			} else if cdrsCount == 21 {
				totalMsat := int64(2100000)
				confirmationPriceMsat, confirmationCommissionMsat, confirmationTaxMsat := session.ReverseCommissionInt64(totalMsat, sessionUser.CommissionPercent, taxPercent)

				invoiceParams := util.InvoiceParams{
					PriceMsat:      dbUtil.SqlNullInt64(confirmationPriceMsat),
					CommissionMsat: dbUtil.SqlNullInt64(confirmationCommissionMsat),
					TaxMsat:        dbUtil.SqlNullInt64(confirmationTaxMsat),
					TotalMsat:      dbUtil.SqlNullInt64(totalMsat),
				}

				r.IssueInvoiceRequest(ctx, sessionUser.ID, &sess.ID, "21_CHARGES", sess.Currency, "Satimoto: 21", invoiceParams)
			}
		}

		// Issue invoice request to circuit user
		circuitPercent := dbUtil.GetEnvFloat64("CIRCUIT_PERCENT", 0.5)
		circuitAmountFiat := (cdrTotalFiat / 100.0) * circuitPercent

		if sessionUser.CircuitUserID.Valid && circuitAmountFiat > 0.0 {
			// TODO: This should be launched as a goroutine to force completion/retries
			releaseDate := time.Now().Add(time.Duration(rand.Intn(120)) * time.Minute)
			invoiceParams := util.InvoiceParams{
				TotalFiat:   dbUtil.SqlNullFloat64(circuitAmountFiat),
				ReleaseDate: dbUtil.SqlNullTime(releaseDate),
			}

			_, err := r.IssueInvoiceRequest(ctx, sessionUser.CircuitUserID.Int64, nil, "CIRCUIT", sess.Currency, "Satimoto: Recharge", invoiceParams)

			return err
		}
	}

	return nil
}
