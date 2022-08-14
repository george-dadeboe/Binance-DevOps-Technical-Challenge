package internal

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	PriceSpread = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "symbol_price_spread",
			Help: "The price spread for a given symbol",
		},
		[]string{"symbol", "price_spread", "delta"},
	)
)
