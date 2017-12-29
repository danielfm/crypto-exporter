package collector

/*
 * This exchange only supports BTC-BRL pair.
 */

import (
	"time"

	"github.com/golang/glog"
	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
	"github.com/prometheus/client_golang/prometheus"
)

// Message sent by the server when an order is created.
type OrderMessage struct {
	Type      int     `json:"type"`
	UnitPrice float64 `json:"unit_price"`
	Amount    float64 `json:"amount"`
}

// Message sent by the server when an order is executed.
type OrderCompletedMessage struct {
	CreateDate time.Time `json:"create_date"`
	Type       int       `json:"type"`
	Amount     float64   `json:"amount"`
	UnitPrice  float64   `json:"unit_price"`
}

// Message sent by the server when an order is canceled.
type CancelOrderMessage struct {
	Type      int     `json:"type"`
	Amount    float64 `json:"amount"`
	UnitPrice float64 `json:"unit_price"`
}

// Message sent periodically by the server with the latest market
// summary information.
type MarketSummaryMessage struct {
	UnitPrice24h             float64 `json:"unit_price_24h"`
	Volume24h                float64 `json:"volume_24h"`
	LastTransactionUnitPrice float64 `json:"last_transaction_unit_price"`
	Currency                 string  `json:"currency"`
}

// Normalizes the operation type from int to string.
func bitcointradeTypeToOperation(t int) string {
	switch t {
	case 1:
		return "bid"
	case 2:
		return "ask"
	default:
		return "unknown"
	}
}

// Socket.io-based collector for Bitcointrade.
type BitcointradeCollector struct {
	metrics *CryptoExchangeMetrics
}

// Creates a new instance of the Bitcointrade metrics collector.
func NewBitcointradeCollector(metrics *CryptoExchangeMetrics) *BitcointradeCollector {
	return &BitcointradeCollector{
		metrics: metrics,
	}
}

// Connects to the websocket endpoint
func (c *BitcointradeCollector) Connect() {
	retry := make(chan bool)

	var mustRetry bool

	for {
		cli, err := bitcointradeWebsocketConnect(c, retry)

		// Could not connect, so let's retry
		if err != nil {
			glog.Errorf("Error connecting to endpoint: %v", err)
			goto retry
		}

		// Wait for a disconnection to be signaled
		mustRetry = <-retry

		// Interrupt retry loop
		if !mustRetry {
			break
		}

	retry:
		// Free resources allocated by the previous connection
		if cli != nil {
			cli.Close()
		}

		// Small fixed connection retry back-off
		time.Sleep(3 * time.Second)
	}
}

func bitcointradeWebsocketConnect(c *BitcointradeCollector, retry chan bool) (*gosocketio.Client, error) {
	tr := transport.GetDefaultWebsocketTransport()

	// Ping intervals based on manual inspection
	tr.PingInterval = 10 * time.Second
	tr.PingTimeout = 5 * time.Second

	cli, err := gosocketio.Dial("wss://core.bitcointrade.com.br/socket.io/?EIO=3&transport=websocket", tr)
	if err != nil {
		return cli, err
	}

	err = cli.On(gosocketio.OnConnection, func(h *gosocketio.Channel) {
		glog.V(2).Infof("Successfully connected to the websocket endpoint")
	})
	if err != nil {
		glog.Fatalf("Cannot listen for connection messages from websocket endpoint: %v", err)
	}

	// Handles disconnect messages
	err = cli.On(gosocketio.OnDisconnection, func(h *gosocketio.Channel) {
		glog.Warningf("Disconnected from websocket endpoint, reconnecting")
		retry <- true
	})
	if err != nil {
		glog.Fatalf("Cannot listen for disconnection messages from websocket endpoint: %v", err)
	}

	// Order metrics
	err = cli.On("order", func(h *gosocketio.Channel, order OrderMessage) {
		glog.V(2).Infof("Received order message: %+v", order)
		c.metrics.orderCount.WithLabelValues("BTC", "BRL", "bitcointrade", bitcointradeTypeToOperation(order.Type), "create").Inc()
	})
	if err != nil {
		glog.Fatalf("Cannot listen for 'order' messages from websocket endpoint: %v", err)
	}

	err = cli.On("cancel_order", func(h *gosocketio.Channel, order CancelOrderMessage) {
		glog.V(2).Infof("Received cancel order message: %+v", order)
		c.metrics.orderCount.WithLabelValues("BTC", "BRL", "bitcointrade", bitcointradeTypeToOperation(order.Type), "cancel").Inc()
	})
	if err != nil {
		glog.Fatalf("Cannot listen for 'cancel_order' messages from websocket endpoint: %v", err)
	}

	// Trade metrics
	err = cli.On("order_completed", func(h *gosocketio.Channel, order OrderCompletedMessage) {
		glog.V(2).Infof("Received order completed message: %+v", order)

		operation := bitcointradeTypeToOperation(order.Type)
		c.metrics.tradeCount.WithLabelValues("BTC", "BRL", "bitcointrade", operation).Inc()
		c.metrics.tradePrice.WithLabelValues("BTC", "BRL", "bitcointrade", operation).Set(order.UnitPrice)
		c.metrics.tradeAmount.WithLabelValues("BTC", "BRL", "bitcointrade", operation).Set(order.Amount)
		c.metrics.tradeAmountSum.WithLabelValues("BTC", "BRL", "bitcointrade", operation).Add(order.Amount)
	})
	if err != nil {
		glog.Fatalf("Cannot listen for 'order_completed' messages from websocket endpoint: %v", err)
	}

	return cli, nil
}

// Returns all available metrics for this collector.
func (c *BitcointradeCollector) Describe(ch chan<- *prometheus.Desc) {
	c.metrics.Describe(ch)
}

// Returns current telemetry data for this collector.
func (c *BitcointradeCollector) Collect(ch chan<- prometheus.Metric) {
	c.metrics.Collect(ch)
}
