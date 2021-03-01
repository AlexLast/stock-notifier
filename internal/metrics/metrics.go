package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// SuccessfulFetches is a counter for successful fetches
	// from a retailer
	SuccessfulFetches = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "stock_notifier_successful_fetches_total",
			Help: "Number of successful fetches from a retailer",
		},
		[]string{
			"retailer",
		},
	)
	// FailedFetches is a counter for failed fetches
	// from a retailer
	FailedFetches = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "stock_notifier_failed_fetches_total",
			Help: "Number of failed fetches from a retailer",
		},
		[]string{
			"retailer",
		},
	)
	// ParsedProducts is a counter for products parsed
	ParsedProducts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "stock_notifier_parsed_products_total",
			Help: "Number of products parsed",
		},
		[]string{
			"retailer",
		},
	)
)
