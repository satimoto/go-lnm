package cdr

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/satimoto/go-datastore/pkg/db"
	"github.com/satimoto/go-datastore/pkg/param"
	dbUtil "github.com/satimoto/go-datastore/pkg/util"
	metrics "github.com/satimoto/go-lsp/internal/metric"
	"github.com/satimoto/go-lsp/internal/notification"
	"github.com/satimoto/go-lsp/pkg/util"
)

func (r *CdrResolver) IssueInvoiceRequest(ctx context.Context, userID int64, promotionCode string, currency string, memo string, invoiceParams util.InvoiceParams) (*db.InvoiceRequest, error) {
	currencyRate, err := r.FerpService.GetRate(currency)

	if err != nil {
		metrics.RecordError("LSP111", "Error retrieving exchange rate", err)
		log.Printf("LSP111: Currency=%v", currency)
		return nil, errors.New("error retrieving exchange rate")
	}

	user, err := r.SessionResolver.UserResolver.Repository.GetUser(ctx, userID)

	if err != nil {
		metrics.RecordError("LSP112", "Error retrieving user", err)
		log.Printf("LSP112: UserID=%v", userID)
		return nil, errors.New("error retrieving user")
	}

	promotion, err := r.PromotionRepository.GetPromotionByCode(ctx, promotionCode)

	if err != nil {
		metrics.RecordError("LSP113", "Error retrieving promotion", err)
		log.Printf("LSP113: Code=%v", promotionCode)
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
		updateInvoiceRequestParams.PriceFiat = addNullFloat64(updateInvoiceRequestParams.PriceFiat, invoiceParams.PriceFiat)
		updateInvoiceRequestParams.PriceMsat = addNullInt64(updateInvoiceRequestParams.PriceMsat, invoiceParams.PriceMsat)
		updateInvoiceRequestParams.CommissionFiat = addNullFloat64(updateInvoiceRequestParams.CommissionFiat, invoiceParams.CommissionFiat)
		updateInvoiceRequestParams.CommissionMsat = addNullInt64(updateInvoiceRequestParams.CommissionMsat, invoiceParams.CommissionMsat)
		updateInvoiceRequestParams.TaxFiat = addNullFloat64(updateInvoiceRequestParams.TaxFiat, invoiceParams.TaxFiat)
		updateInvoiceRequestParams.TaxMsat = addNullInt64(updateInvoiceRequestParams.TaxMsat, invoiceParams.TaxMsat)
		updateInvoiceRequestParams.TotalFiat = updateInvoiceRequestParams.TotalFiat + invoiceParams.TotalFiat.Float64
		updateInvoiceRequestParams.TotalMsat = updateInvoiceRequestParams.TotalMsat + invoiceParams.TotalMsat.Int64

		invoiceRequest, err = r.InvoiceRequestRepository.UpdateInvoiceRequest(ctx, updateInvoiceRequestParams)

		if err != nil {
			metrics.RecordError("LSP114", "Error updating invoice request", err)
			log.Printf("LSP114: Params=%#v", updateInvoiceRequestParams)
			return nil, errors.New("error updating invoice request")
		}
	} else {
		createInvoiceRequestParams := db.CreateInvoiceRequestParams{
			UserID:         user.ID,
			PromotionID:    promotion.ID,
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
			metrics.RecordError("LSP115", "Error creating invoice request", err)
			log.Printf("LSP115: Params=%#v", createInvoiceRequestParams)
			return nil, errors.New("error creating invoice request")
		}

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
			metrics.RecordError("LSP130", "Error creating pending notification", err)
			log.Printf("LSP130: Params=%#v", createPendingNotificationParams)
			return nil, errors.New("error creating pending notification")
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

func addNullFloat64(floatA, floatB sql.NullFloat64) sql.NullFloat64 {
	return dbUtil.SqlNullFloat64(floatA.Float64 + floatB.Float64)
}

func addNullInt64(intA, intB sql.NullInt64) sql.NullInt64 {
	return dbUtil.SqlNullInt64(intA.Int64 + intB.Int64)
}
