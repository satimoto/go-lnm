package util

import (
	"database/sql"

	dbUtil "github.com/satimoto/go-datastore/pkg/util"
)

type InvoiceParams struct {
	PriceFiat      sql.NullFloat64
	PriceMsat      sql.NullInt64
	CommissionFiat sql.NullFloat64
	CommissionMsat sql.NullInt64
	TaxFiat        sql.NullFloat64
	TaxMsat        sql.NullInt64
	TotalFiat      sql.NullFloat64
	TotalMsat      sql.NullInt64
}

func FillInvoiceRequestParams(invoiceParams InvoiceParams, rateMsat float64) InvoiceParams {
	invoiceParams.PriceFiat, invoiceParams.PriceMsat = fillInvoiceRequestParam(invoiceParams.PriceFiat, invoiceParams.PriceMsat, rateMsat)
	invoiceParams.CommissionFiat, invoiceParams.CommissionMsat = fillInvoiceRequestParam(invoiceParams.CommissionFiat, invoiceParams.CommissionMsat, rateMsat)
	invoiceParams.TaxFiat, invoiceParams.TaxMsat = fillInvoiceRequestParam(invoiceParams.TaxFiat, invoiceParams.TaxMsat, rateMsat)
	invoiceParams.TotalFiat, invoiceParams.TotalMsat = fillInvoiceRequestParam(invoiceParams.TotalFiat, invoiceParams.TotalMsat, rateMsat)

	return invoiceParams
}

func fillInvoiceRequestParam(amountFiat sql.NullFloat64, amountMsat sql.NullInt64, rateMsat float64) (sql.NullFloat64, sql.NullInt64) {
	if amountFiat.Valid || amountMsat.Valid {
		switch {
		case !amountFiat.Valid:
			amountFiat = dbUtil.SqlNullFloat64(float64(amountMsat.Int64) / rateMsat)
		case !amountMsat.Valid:
			amountMsat = dbUtil.SqlNullInt64(int64(amountFiat.Float64 * rateMsat))
		}
	}

	return amountFiat, amountMsat
}
