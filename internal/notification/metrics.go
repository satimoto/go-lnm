package notification

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricNotificationsSentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "lsp_notifications_sent_total",
		Help: "The total number of notifications sent",
	}, []string{"type"})
)

func RecordNotificationSent(notifcationType string, count int) {
	metricNotificationsSentTotal.WithLabelValues(notifcationType).Add(float64(count))
}
