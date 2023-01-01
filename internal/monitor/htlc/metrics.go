package htlc

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricInterceptedChannelRequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lsp_intercepted_channel_requests_total",
		Help: "The total number of intercepted channel requests",
	})
	metricInterceptedChannelRequestsSettled = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lsp_intercepted_channel_requests_settled",
		Help: "The total number of intercepted channel requests settled",
	})
	metricInterceptedChannelRequestsFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lsp_intercepted_channel_requests_failed",
		Help: "The total number of intercepted channel requests failed",
	})
)
