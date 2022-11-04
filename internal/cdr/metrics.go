package cdr

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricCdrsFlaggedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lsp_cdrs_flagged_total",
		Help: "The total number of cdrs flagged",
	})
	metricInvoiceRequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lsp_invoice_requests_total",
		Help: "The total number of invoice requests",
	})
	metricInvoiceRequestsCommissionFiat = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "lsp_invoice_requests_commission_fiat",
		Help: "The total commission payment in fiat",
	}, []string{"currency"})
	metricInvoiceRequestsCommissionSatoshis = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lsp_invoice_requests_commission_satoshis",
		Help: "The total commission payment in satoshis",
	})
	metricInvoiceRequestsPriceFiat = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "lsp_invoice_requests_price_fiat",
		Help: "The total price payment in fiat",
	}, []string{"currency"})
	metricInvoiceRequestsPriceSatoshis = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lsp_invoice_requests_price_satoshis",
		Help: "The total price payment in satoshis",
	})
	metricInvoiceRequestsTaxFiat = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "lsp_invoice_requests_tax_fiat",
		Help: "The total tax payment in fiat",
	}, []string{"currency"})
	metricInvoiceRequestsTaxSatoshis = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lsp_invoice_requests_tax_satoshis",
		Help: "The total tax payment in satoshis",
	})
	metricInvoiceRequestsTotalFiat = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "lsp_invoice_requests_total_fiat",
		Help: "The total payment in fiat",
	}, []string{"currency"})
	metricInvoiceRequestsTotalSatoshis = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lsp_invoice_requests_total_satoshis",
		Help: "The total payment in satoshis",
	})
)
