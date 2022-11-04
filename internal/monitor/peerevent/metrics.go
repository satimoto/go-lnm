package peerevent

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricPeersOnlineTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "lsp_peers_online_total",
		Help: "The total number of peers online",
	})
)

func RecordPeersOnline(count uint32) {
	metricPeersOnlineTotal.Set(float64(count))
}
