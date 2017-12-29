package main

import (
	"flag"
	"net/http"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/danielfm/crypto-exporter/collector"
)

var (
	// VERSION set by build script
	VERSION = "UNKNOWN"

	addr             = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	metricsEndpoint  = flag.String("endpoint", "/metrics", "Path under which to expose metrics.")
	metricsNamespace = flag.String("namespace", "crypto", "Metrics namespace.")
)

func init() {
	flag.Parse()

	// TODO: always log to stderr for now
	flag.Set("logtostderr", "true")
}

func main() {
	glog.Infof("Crypto Exporter v%s started, listening on %s.", VERSION, *addr)
	glog.Infof("Parameters: endpoint=%s, namespace=%s", *metricsEndpoint, *metricsNamespace)

	metrics := collector.NewCryptoExchangeMetrics(*metricsNamespace)

	// Registers collectors for each supported exchange
	bitcointradeCollector := collector.NewBitcointradeCollector(metrics)
	prometheus.Register(bitcointradeCollector)

	// Stream metrics from each supported exchange in background
	go bitcointradeCollector.Connect()

	http.Handle(*metricsEndpoint, promhttp.Handler())
	glog.Fatal(http.ListenAndServe(*addr, nil))
}
