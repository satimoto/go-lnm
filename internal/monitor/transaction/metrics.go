package transaction

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricWalletTotalBalanceSatoshis = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "lsp_wallet_total_balance_satoshis",
		Help: "The total wallet balance in satoshis",
	})
	metricWalletConfirmedBalanceSatoshis = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "lsp_wallet_confirmed_balance_satoshis",
		Help: "The confirmed wallet balance in satoshis",
	})
	metricWalletUnconfirmedBalanceSatoshis = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "lsp_wallet_unconfirmed_balance_satoshis",
		Help: "The unconfirmed wallet balance in satoshis",
	})
	metricWalletLockedBalanceSatoshis = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "lsp_wallet_locked_balance_satoshis",
		Help: "The locked wallet balance in satoshis",
	})
	metricWalletReservedBalanceSatoshis = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "lsp_wallet_reserved_balance_satoshis",
		Help: "The reserved wallet balance in satoshis",
	})
)
