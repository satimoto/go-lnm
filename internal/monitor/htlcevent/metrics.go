package htlcevent

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricRoutingEventsSettled = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lsp_routing_events_settled",
		Help: "The total number of routing events settled",
	})
	metricRoutingEventsSettledFeeFiat = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "lsp_routing_events_settled_fee_fiat",
		Help: "The total fees received in fiat",
	}, []string{"currency"})
	metricRoutingEventsSettledFeeSatoshis = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lsp_routing_events_settled_fee_satoshis",
		Help: "The total fees received in satoshis",
	})
	metricRoutingEventsSettledTotalFiat = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "lsp_routing_events_settled_total_fiat",
		Help: "The total routed in fiat",
	}, []string{"currency"})
	metricRoutingEventsSettledTotalSatoshis = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lsp_routing_events_settled_total_satoshis",
		Help: "The total routed in satoshis",
	})

	metricRoutingEventsFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lsp_routing_events_failed",
		Help: "The total number of routing events failed",
	})
	metricRoutingEventsFailedFeeFiat = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "lsp_routing_events_failed_fee_fiat",
		Help: "The total fees missed in fiat",
	}, []string{"currency"})
	metricRoutingEventsFailedFeeSatoshis = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lsp_routing_events_failed_fee_satoshis",
		Help: "The total fees missed in satoshis",
	})
	metricRoutingEventsFailedTotalFiat = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "lsp_routing_events_failed_total_fiat",
		Help: "The total failed in fiat",
	}, []string{"currency"})
	metricRoutingEventsFailedTotalSatoshis = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lsp_routing_events_failed_total_satoshis",
		Help: "The total failed in satoshis",
	})
)
