package invoice

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricSessionInvoicesSettledTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lsp_session_invoices_settled_total",
		Help: "The total number of session invoices settled",
	})
)
