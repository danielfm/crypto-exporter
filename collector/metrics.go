package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Gathers metrics across all supported crypto exchanges. Depending on the
// provided API, some metrics might not be available in some particular
// exchange.
type CryptoExchangeMetrics struct {
	// Order book metrics
	orderCount *prometheus.CounterVec

	// Completed order metrics
	tradeCount     *prometheus.CounterVec
	tradePrice     *prometheus.GaugeVec
	tradeAmount    *prometheus.GaugeVec
	tradeAmountSum *prometheus.CounterVec
}

// Returns a new instance of CryptoExchangeMetrics scoped to a particular
// namespace.
func NewCryptoExchangeMetrics(namespace string) *CryptoExchangeMetrics {
	orderLabels := []string{"base_currency", "quote_currency", "exchange_name", "operation", "event"}
	tradeLabels := []string{"base_currency", "quote_currency", "exchange_name", "operation"}

	return &CryptoExchangeMetrics{
		// Order book metrics
		orderCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: prometheus.BuildFQName(namespace, "order", "count"),
				Help: "Number of unexecuted/canceled orders.",
			}, orderLabels,
		),

		// Completed order metrics
		tradeCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: prometheus.BuildFQName(namespace, "trade", "count"),
				Help: "Number of executed orders.",
			}, tradeLabels,
		),
		tradePrice: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(namespace, "trade", "price"),
				Help: "Last trade amount, in quote currency.",
			}, tradeLabels,
		),
		tradeAmount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(namespace, "trade", "amount"),
				Help: "Last trade amount, in base currency.",
			}, tradeLabels,
		),
		tradeAmountSum: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: prometheus.BuildFQName(namespace, "trade", "amount_sum"),
				Help: "Sum of all trade amounts, in base currency.",
			}, tradeLabels,
		),
	}
}

// Returns all available metrics for the metrics exposed by this type.
func (c *CryptoExchangeMetrics) Describe(ch chan<- *prometheus.Desc) {
	// Order book metrics
	c.orderCount.Describe(ch)

	// Completed order metrics
	c.tradeCount.Describe(ch)
	c.tradePrice.Describe(ch)
	c.tradeAmount.Describe(ch)
	c.tradeAmountSum.Describe(ch)
}

// Returns current telemetry data for the metrics exposed by this type.
func (c *CryptoExchangeMetrics) Collect(ch chan<- prometheus.Metric) {
	// Order book metrics
	c.orderCount.Collect(ch)

	// Completed order metrics
	c.tradeCount.Collect(ch)
	c.tradePrice.Collect(ch)
	c.tradeAmount.Collect(ch)
	c.tradeAmountSum.Collect(ch)
}
