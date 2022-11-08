package session

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricSessionsFlaggedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lsp_sessions_flagged_total",
		Help: "The total number of sessions flagged",
	})
	metricSessionMonitoringGoroutines = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "lsp_session_monitoring_goroutines",
		Help: "The total number of session monitoring goroutines",
	})
	metricSessionInvoicesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lsp_session_invoices_total",
		Help: "The total number of session invoices",
	})
	metricSessionInvoicesCommissionFiat = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "lsp_session_invoices_commission_fiat",
		Help: "The total commission invoiced in fiat",
	}, []string{"currency"})
	metricSessionInvoicesCommissionSatoshis = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lsp_session_invoices_commission_satoshis",
		Help: "The total commission invoiced in satoshis",
	})
	metricSessionInvoicesPriceFiat = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "lsp_session_invoices_price_fiat",
		Help: "The total price invoiced in fiat",
	}, []string{"currency"})
	metricSessionInvoicesPriceSatoshis = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lsp_session_invoices_price_satoshis",
		Help: "The total price invoiced in satoshis",
	})
	metricSessionInvoicesTaxFiat = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "lsp_session_invoices_tax_fiat",
		Help: "The total tax invoiced in fiat",
	}, []string{"currency"})
	metricSessionInvoicesTaxSatoshis = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lsp_session_invoices_tax_satoshis",
		Help: "The total tax invoiced in satoshis",
	})
	metricSessionInvoicesTotalFiat = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "lsp_session_invoices_total_fiat",
		Help: "The total invoiced in fiat",
	}, []string{"currency"})
	metricSessionInvoicesTotalSatoshis = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lsp_session_invoices_total_satoshis",
		Help: "The total invoiced in satoshis",
	})
)

func RecordFlaggedSession() {
	metricSessionsFlaggedTotal.Inc()
}