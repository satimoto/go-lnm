package channelevent

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricChannelsTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "lsp_channels_total",
		Help: "The total number of channels",
	})
)

func RecordChannels(count uint32) {
	metricChannelsTotal.Set(float64(count))
}
