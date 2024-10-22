package session

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/param"
	dbUtil "github.com/satimoto/go-datastore/pkg/util"
	"github.com/satimoto/go-ferp/pkg/rate"
	"github.com/satimoto/go-lnm/internal/lightningnetwork"
	metrics "github.com/satimoto/go-lnm/internal/metric"
	"github.com/satimoto/go-lnm/pkg/util"
)

func (r *SessionResolver) IssueSessionInvoice(ctx context.Context, user db.User, session db.Session, invoiceParams util.InvoiceParams, chargeParams util.ChargeParams) *db.SessionInvoice {
	currencyRate, err := r.FerpService.GetRate(invoiceParams.Currency)

	if err != nil {
		metrics.RecordError("LNM054", "Error retrieving exchange rate", err)
		log.Printf("LNM054: Currency=%v", invoiceParams.Currency)
		return nil
	}

	updateUnsettledInvoices := dbUtil.GetEnvBool("UPDATE_UNSETTLED_INVOICES", false)

	if updateUnsettledInvoices {
		if sessionInvoice, err := r.Repository.GetUnsettledSessionInvoiceBySession(ctx, session.ID); err == nil {
			// An unsettled session invoice exists, try to update it
			if updatedSessionInvoice := r.updateSessionInvoice(ctx, currencyRate, session, sessionInvoice, invoiceParams, chargeParams); updatedSessionInvoice != nil {
				return updatedSessionInvoice
			}
		}
	}

	// Create a new session invoice
	return r.createSessionInvoice(ctx, currencyRate, user, session, invoiceParams, chargeParams)
}

func (r *SessionResolver) WaitForInvoiceExpiry(paymentRequest string) {
	expiry := int64(3600)
	payReqParams := &lnrpc.PayReqString{PayReq: paymentRequest}
	payReqResponse, err := r.LightningService.DecodePayReq(payReqParams)

	if err == nil {
		expiry = payReqResponse.Expiry
	}

	ctx := context.Background()
	timeout := (time.Second * time.Duration(expiry)) - (time.Second * 90)

	time.Sleep(timeout)

	if sessionInvoice, err := r.Repository.GetSessionInvoiceByPaymentRequest(ctx, paymentRequest); err == nil && !sessionInvoice.IsSettled {
		if paymentRequest, signature, err := lightningnetwork.CreateLightningInvoice(r.LightningService, payReqResponse.Description, sessionInvoice.TotalMsat); err == nil {
			// Get the session invoice again to check if it's been settled or updated
			latestSessionInvoice, err := r.Repository.GetSessionInvoice(ctx, sessionInvoice.ID)

			if err == nil && !latestSessionInvoice.IsSettled && latestSessionInvoice.PaymentRequest == sessionInvoice.PaymentRequest {
				sessionInvoiceParams := param.NewUpdateSessionInvoiceParams(sessionInvoice)
				sessionInvoiceParams.PaymentRequest = paymentRequest
				sessionInvoiceParams.Signature = signature
				sessionInvoiceParams.IsExpired = false

				_, err := r.Repository.UpdateSessionInvoice(ctx, sessionInvoiceParams)

				if err != nil {
					metrics.RecordError("LNM170", "Error updating session invoice", err)
					log.Printf("LNM170: Params=%#v", sessionInvoiceParams)
				}

				go r.WaitForInvoiceExpiry(paymentRequest)
			}

			return
		}

		updateSessionInvoiceParams := param.NewUpdateSessionInvoiceParams(sessionInvoice)
		updateSessionInvoiceParams.IsExpired = true

		_, err = r.Repository.UpdateSessionInvoice(ctx, updateSessionInvoiceParams)

		if err != nil {
			metrics.RecordError("LNM036", "Error updating session invoice", err)
			log.Printf("LNM036: Params=%#v", updateSessionInvoiceParams)
		}

		// Metrics: Increment number of expired session invoices
		metricSessionInvoicesExpiredTotal.Inc()
	}
}

func (r *SessionResolver) createSessionInvoice(ctx context.Context, currencyRate *rate.CurrencyRate, user db.User, session db.Session, invoiceParams util.InvoiceParams, chargeParams util.ChargeParams) *db.SessionInvoice {
	memo := fmt.Sprintf("Satimoto: %s", session.Uid)
	rateMsat := float64(currencyRate.RateMsat)
	invoiceParams = util.FillInvoiceRequestParams(invoiceParams, rateMsat)

	if !invoiceParams.TotalMsat.Valid {
		metrics.RecordError("LNM116", "Error filling request params", errors.New("invoiceParams TotalMsat not valid"))
		log.Printf("LNM116: SessionUid=%v, Params=%#v", session.Uid, invoiceParams)
		return nil
	}

	if paymentRequest, signature, err := lightningnetwork.CreateLightningInvoice(r.LightningService, memo, invoiceParams.TotalMsat.Int64); err == nil {
		sessionInvoiceParams := param.NewCreateSessionInvoiceParams(session)
		sessionInvoiceParams.UserID = user.ID
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
		sessionInvoiceParams.EstimatedEnergy = chargeParams.EstimatedEnergy
		sessionInvoiceParams.EstimatedTime = chargeParams.EstimatedTime
		sessionInvoiceParams.MeteredEnergy = chargeParams.MeteredEnergy
		sessionInvoiceParams.MeteredTime = chargeParams.MeteredTime

		sessionInvoice, err := r.Repository.CreateSessionInvoice(ctx, sessionInvoiceParams)

		if err != nil {
			metrics.RecordError("LNM003", "Error creating session invoice", err)
			log.Printf("LNM003: Params=%#v", sessionInvoiceParams)
			return nil
		}

		// Metrics
		metricSessionInvoicesTotal.Inc()
		metricSessionInvoicesCommissionFiat.WithLabelValues(invoiceParams.Currency).Add(sessionInvoice.CommissionFiat)
		metricSessionInvoicesCommissionSatoshis.Add(float64(sessionInvoice.CommissionMsat / 1000))
		metricSessionInvoicesPriceFiat.WithLabelValues(invoiceParams.Currency).Add(sessionInvoice.PriceFiat)
		metricSessionInvoicesPriceSatoshis.Add(float64(sessionInvoice.PriceMsat / 1000))
		metricSessionInvoicesTaxFiat.WithLabelValues(invoiceParams.Currency).Add(sessionInvoice.TaxFiat)
		metricSessionInvoicesTaxSatoshis.Add(float64(sessionInvoice.TaxMsat / 1000))
		metricSessionInvoicesTotalFiat.WithLabelValues(invoiceParams.Currency).Add(sessionInvoice.TotalFiat)
		metricSessionInvoicesTotalSatoshis.Add(float64(sessionInvoice.TotalMsat / 1000))

		// TODO: handle notification failure
		r.SendSessionInvoiceNotification(user, session, sessionInvoice)

		go r.WaitForInvoiceExpiry(paymentRequest)

		return &sessionInvoice
	}

	return nil
}

func (r *SessionResolver) updateSessionInvoice(ctx context.Context, currencyRate *rate.CurrencyRate, session db.Session, sessionInvoice db.SessionInvoice, invoiceParams util.InvoiceParams, chargeParams util.ChargeParams) *db.SessionInvoice {
	memo := fmt.Sprintf("Satimoto: %s", session.Uid)
	rateMsat := float64(currencyRate.RateMsat)
	updateInvoiceParams := util.InvoiceParams{
		Currency:       invoiceParams.Currency,
		PriceFiat:      util.AddNullFloat64(dbUtil.SqlNullFloat64(sessionInvoice.PriceFiat), invoiceParams.PriceFiat),
		CommissionFiat: util.AddNullFloat64(dbUtil.SqlNullFloat64(sessionInvoice.CommissionFiat), invoiceParams.CommissionFiat),
		TaxFiat:        util.AddNullFloat64(dbUtil.SqlNullFloat64(sessionInvoice.TaxFiat), invoiceParams.TaxFiat),
		TotalFiat:      util.AddNullFloat64(dbUtil.SqlNullFloat64(sessionInvoice.TotalFiat), invoiceParams.TotalFiat),
	}

	updateInvoiceParams = util.FillInvoiceRequestParams(updateInvoiceParams, rateMsat)

	if !invoiceParams.TotalMsat.Valid {
		metrics.RecordError("LNM168", "Error filling request params", errors.New("invoiceParams TotalMsat not valid"))
		log.Printf("LNM168: SessionInvoiceID=%v, Params=%#v", sessionInvoice.ID, invoiceParams)
		return nil
	}

	if paymentRequest, signature, err := lightningnetwork.CreateLightningInvoice(r.LightningService, memo, invoiceParams.TotalMsat.Int64); err == nil {
		// Get the session invoice again to check if it's been settled or updated
		latestSessionInvoice, err := r.Repository.GetSessionInvoice(ctx, sessionInvoice.ID)

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

			_, err := r.Repository.UpdateSessionInvoice(ctx, sessionInvoiceParams)

			if err != nil {
				metrics.RecordError("LNM169", "Error updating session invoice", err)
				log.Printf("LNM169: Params=%#v", sessionInvoiceParams)
				return nil
			}

			// Metrics
			metricSessionInvoicesTotal.Inc()
			metricSessionInvoicesCommissionFiat.WithLabelValues(invoiceParams.Currency).Add(invoiceParams.CommissionFiat.Float64)
			metricSessionInvoicesCommissionSatoshis.Add(float64(invoiceParams.CommissionMsat.Int64 / 1000))
			metricSessionInvoicesPriceFiat.WithLabelValues(invoiceParams.Currency).Add(invoiceParams.PriceFiat.Float64)
			metricSessionInvoicesPriceSatoshis.Add(float64(invoiceParams.PriceMsat.Int64 / 1000))
			metricSessionInvoicesTaxFiat.WithLabelValues(invoiceParams.Currency).Add(invoiceParams.TaxFiat.Float64)
			metricSessionInvoicesTaxSatoshis.Add(float64(invoiceParams.TaxMsat.Int64 / 1000))
			metricSessionInvoicesTotalFiat.WithLabelValues(invoiceParams.Currency).Add(invoiceParams.TotalFiat.Float64)
			metricSessionInvoicesTotalSatoshis.Add(float64(invoiceParams.TotalMsat.Int64 / 1000))

			go r.WaitForInvoiceExpiry(paymentRequest)

			return &sessionInvoice
		}
	}

	return nil
}
