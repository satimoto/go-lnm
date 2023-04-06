package cdr

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/param"
	dbUtil "github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-lnm/internal/lightningnetwork"
	metrics "github.com/satimoto/go-lnm/internal/metric"
	"github.com/satimoto/go-lnm/internal/notification"
	"github.com/satimoto/go-lnm/pkg/util"
)

func (r *CdrResolver) IssueRebate(ctx context.Context, session db.Session, userID int64, invoiceParams util.InvoiceParams, chargeParams util.ChargeParams) {
	updateUnsettledInvoices := dbUtil.GetEnvBool("UPDATE_UNSETTLED_INVOICES", false)

	if updateUnsettledInvoices {
		sessionInvoice, err := r.SessionResolver.Repository.GetUnsettledSessionInvoiceBySession(ctx, session.ID)

		// First try to rebute via deducting from an unsettled session invoice
		if err == nil && invoiceParams.TotalFiat.Valid && sessionInvoice.TotalFiat > invoiceParams.TotalFiat.Float64 {
			// An unsettled session invoice exists, try to update it
			if updatedSessionInvoice := r.updateSessionInvoice(ctx, session, sessionInvoice, invoiceParams, chargeParams); updatedSessionInvoice != nil {
				return
			}
		}
	}

	// Then issue an invoice request if no session invoice exists
	memo := fmt.Sprintf("Satimoto: %s", session.Uid)

	if invoiceRequest, err := r.IssueInvoiceRequest(ctx, userID, &session.ID, "REBATE", session.Currency, memo, invoiceParams); err == nil {
		updateSessionByUidParams := param.NewUpdateSessionByUidParams(session)
		updateSessionByUidParams.InvoiceRequestID = dbUtil.SqlNullInt64(invoiceRequest.ID)

		_, err := r.SessionResolver.Repository.UpdateSessionByUid(ctx, updateSessionByUidParams)

		if err != nil {
			metrics.RecordError("LNM117", "Error updating session", err)
			log.Printf("LNM117: Params=%v", updateSessionByUidParams)
		}
	}
}

func (r *CdrResolver) IssueInvoiceRequest(ctx context.Context, userID int64, sessionID *int64, promotionCode string, currency string, memo string, invoiceParams util.InvoiceParams) (*db.InvoiceRequest, error) {
	currencyRate, err := r.FerpService.GetRate(currency)

	if err != nil {
		metrics.RecordError("LNM111", "Error retrieving exchange rate", err)
		log.Printf("LNM111: Currency=%v", currency)
		return nil, errors.New("error retrieving exchange rate")
	}

	user, err := r.SessionResolver.UserResolver.Repository.GetUser(ctx, userID)

	if err != nil {
		metrics.RecordError("LNM112", "Error retrieving user", err)
		log.Printf("LNM112: UserID=%v", userID)
		return nil, errors.New("error retrieving user")
	}

	promotion, err := r.PromotionRepository.GetPromotionByCode(ctx, promotionCode)

	if err != nil {
		metrics.RecordError("LNM113", "Error retrieving promotion", err)
		log.Printf("LNM113: Code=%v", promotionCode)
		return nil, errors.New("error retrieving promotion")
	}

	rateMsat := float64(currencyRate.RateMsat)
	invoiceParams = util.FillInvoiceRequestParams(invoiceParams, rateMsat)

	getUnsettledInvoiceRequestParams := db.GetUnsettledInvoiceRequestParams{
		UserID:      user.ID,
		PromotionID: promotion.ID,
		Memo:        memo,
	}

	invoiceRequest, err := r.InvoiceRequestRepository.GetUnsettledInvoiceRequest(ctx, getUnsettledInvoiceRequestParams)

	if err == nil {
		updateInvoiceRequestParams := param.NewUpdateInvoiceRequestParams(invoiceRequest)
		updateInvoiceRequestParams.PriceFiat = util.AddNullFloat64(updateInvoiceRequestParams.PriceFiat, invoiceParams.PriceFiat)
		updateInvoiceRequestParams.PriceMsat = util.AddNullInt64(updateInvoiceRequestParams.PriceMsat, invoiceParams.PriceMsat)
		updateInvoiceRequestParams.CommissionFiat = util.AddNullFloat64(updateInvoiceRequestParams.CommissionFiat, invoiceParams.CommissionFiat)
		updateInvoiceRequestParams.CommissionMsat = util.AddNullInt64(updateInvoiceRequestParams.CommissionMsat, invoiceParams.CommissionMsat)
		updateInvoiceRequestParams.TaxFiat = util.AddNullFloat64(updateInvoiceRequestParams.TaxFiat, invoiceParams.TaxFiat)
		updateInvoiceRequestParams.TaxMsat = util.AddNullInt64(updateInvoiceRequestParams.TaxMsat, invoiceParams.TaxMsat)
		updateInvoiceRequestParams.TotalFiat = updateInvoiceRequestParams.TotalFiat + invoiceParams.TotalFiat.Float64
		updateInvoiceRequestParams.TotalMsat = updateInvoiceRequestParams.TotalMsat + invoiceParams.TotalMsat.Int64

		invoiceRequest, err = r.InvoiceRequestRepository.UpdateInvoiceRequest(ctx, updateInvoiceRequestParams)

		if err != nil {
			metrics.RecordError("LNM114", "Error updating invoice request", err)
			log.Printf("LNM114: Params=%#v", updateInvoiceRequestParams)
			return nil, errors.New("error updating invoice request")
		}
	} else {
		createInvoiceRequestParams := db.CreateInvoiceRequestParams{
			UserID:         user.ID,
			PromotionID:    promotion.ID,
			SessionID:      dbUtil.SqlNullInt64(sessionID),
			Currency:       currency,
			Memo:           memo,
			PriceFiat:      invoiceParams.PriceFiat,
			PriceMsat:      invoiceParams.PriceMsat,
			CommissionFiat: invoiceParams.CommissionFiat,
			CommissionMsat: invoiceParams.CommissionMsat,
			TaxFiat:        invoiceParams.TaxFiat,
			TaxMsat:        invoiceParams.TaxMsat,
			TotalFiat:      invoiceParams.TotalFiat.Float64,
			TotalMsat:      invoiceParams.TotalMsat.Int64,
			ReleaseDate:    invoiceParams.ReleaseDate,
			IsSettled:      false,
		}

		invoiceRequest, err = r.InvoiceRequestRepository.CreateInvoiceRequest(ctx, createInvoiceRequestParams)

		if err != nil {
			metrics.RecordError("LNM115", "Error creating invoice request", err)
			log.Printf("LNM115: Params=%#v", createInvoiceRequestParams)
			return nil, errors.New("error creating invoice request")
		}

		if invoiceParams.ReleaseDate.Valid {
			// Send a notification 1 day after release date
			sendDate := time.Now()

			if invoiceParams.ReleaseDate.Valid {
				sendDate = invoiceParams.ReleaseDate.Time
			}

			createPendingNotificationParams := db.CreatePendingNotificationParams{
				UserID:           user.ID,
				NodeID:           user.NodeID.Int64,
				InvoiceRequestID: dbUtil.SqlNullInt64(invoiceRequest.ID),
				DeviceToken:      user.DeviceToken,
				Type:             notification.INVOICE_REQUEST,
				SendDate:         sendDate.Add(time.Hour * 24),
			}

			_, err := r.PendingNotificationRepository.CreatePendingNotification(ctx, createPendingNotificationParams)

			if err != nil {
				metrics.RecordError("LNM130", "Error creating pending notification", err)
				log.Printf("LNM130: Params=%#v", createPendingNotificationParams)
				return nil, errors.New("error creating pending notification")
			}
		} else {
			r.SendInvoiceRequestNotification(user, invoiceRequest)
		}

		metricInvoiceRequestsTotal.Inc()
	}

	// Metrics
	metricInvoiceRequestsCommissionFiat.WithLabelValues(currency).Add(invoiceParams.CommissionFiat.Float64)
	metricInvoiceRequestsCommissionSatoshis.Add(float64(invoiceParams.CommissionMsat.Int64 / 1000))
	metricInvoiceRequestsPriceFiat.WithLabelValues(currency).Add(invoiceParams.PriceFiat.Float64)
	metricInvoiceRequestsPriceSatoshis.Add(float64(invoiceParams.PriceMsat.Int64 / 1000))
	metricInvoiceRequestsTaxFiat.WithLabelValues(currency).Add(invoiceParams.TaxFiat.Float64)
	metricInvoiceRequestsTaxSatoshis.Add(float64(invoiceParams.TaxMsat.Int64 / 1000))
	metricInvoiceRequestsTotalFiat.WithLabelValues(currency).Add(invoiceRequest.TotalFiat)
	metricInvoiceRequestsTotalSatoshis.Add(float64(invoiceParams.TotalMsat.Int64 / 1000))

	return &invoiceRequest, nil
}

func (r *CdrResolver) updateSessionInvoice(ctx context.Context, session db.Session, sessionInvoice db.SessionInvoice, invoiceParams util.InvoiceParams, chargeParams util.ChargeParams) *db.SessionInvoice {
	currencyRate, err := r.FerpService.GetRate(session.Currency)

	if err != nil {
		metrics.RecordError("LNM171", "Error retrieving exchange rate", err)
		log.Printf("LNM171: Currency=%v", session.Currency)
		return nil
	}

	rateMsat := float64(currencyRate.RateMsat)
	updateInvoiceParams := util.InvoiceParams{
		PriceFiat:      util.MinusNullFloat64(dbUtil.SqlNullFloat64(sessionInvoice.PriceFiat), invoiceParams.PriceFiat),
		CommissionFiat: util.MinusNullFloat64(dbUtil.SqlNullFloat64(sessionInvoice.CommissionFiat), invoiceParams.CommissionFiat),
		TaxFiat:        util.MinusNullFloat64(dbUtil.SqlNullFloat64(sessionInvoice.TaxFiat), invoiceParams.TaxFiat),
		TotalFiat:      util.MinusNullFloat64(dbUtil.SqlNullFloat64(sessionInvoice.TotalFiat), invoiceParams.TotalFiat),
	}

	memo := fmt.Sprintf("Satimoto: %s", session.Uid)
	updateInvoiceParams = util.FillInvoiceRequestParams(updateInvoiceParams, rateMsat)

	if !invoiceParams.TotalMsat.Valid {
		metrics.RecordError("LNM172", "Error filling request params", errors.New("invoiceParams TotalMsat not valid"))
		log.Printf("LNM172: SessionInvoiceID=%v, Params=%#v", sessionInvoice.ID, invoiceParams)
		return nil
	}

	if paymentRequest, signature, err := lightningnetwork.CreateLightningInvoice(r.LightningService, memo, invoiceParams.TotalMsat.Int64); err == nil {
		// Get the session invoice again to check if it's been settled or updated
		latestSessionInvoice, err := r.SessionResolver.Repository.GetSessionInvoice(ctx, sessionInvoice.ID)

		if err == nil && !latestSessionInvoice.IsSettled && latestSessionInvoice.PaymentRequest == sessionInvoice.PaymentRequest {
			sessionInvoiceParams := param.NewUpdateSessionInvoiceParams(latestSessionInvoice)
			sessionInvoiceParams.CurrencyRate = currencyRate.Rate
			sessionInvoiceParams.CurrencyRateMsat = currencyRate.RateMsat
			sessionInvoiceParams.PriceFiat = invoiceParams.PriceFiat.Float64
			sessionInvoiceParams.PriceMsat = invoiceParams.PriceMsat.Int64
			sessionInvoiceParams.CommissionFiat = invoiceParams.CommissionFiat.Float64
			sessionInvoiceParams.CommissionMsat = invoiceParams.CommissionMsat.Int64
			sessionInvoiceParams.TaxFiat = invoiceParams.TaxFiat.Float64
			sessionInvoiceParams.TaxMsat = invoiceParams.TaxMsat.Int64
			sessionInvoiceParams.TotalFiat = invoiceParams.TotalFiat.Float64
			sessionInvoiceParams.TotalMsat = invoiceParams.TotalMsat.Int64
			sessionInvoiceParams.PaymentRequest = paymentRequest
			sessionInvoiceParams.Signature = signature
			sessionInvoiceParams.IsExpired = false
			sessionInvoiceParams.EstimatedEnergy = chargeParams.EstimatedEnergy
			sessionInvoiceParams.EstimatedTime = chargeParams.EstimatedTime
			sessionInvoiceParams.MeteredEnergy = chargeParams.MeteredEnergy
			sessionInvoiceParams.MeteredTime = chargeParams.MeteredTime

			_, err := r.SessionResolver.Repository.UpdateSessionInvoice(ctx, sessionInvoiceParams)

			if err != nil {
				metrics.RecordError("LNM173", "Error updating session invoice", err)
				log.Printf("LNM173: Params=%#v", sessionInvoiceParams)
				return nil
			}

			go r.SessionResolver.WaitForInvoiceExpiry(paymentRequest)

			return &sessionInvoice
		}
	}

	return nil
}
