package main

import (
	"flag"
	"github.com/jurgen-kluft/go-conbee/sensors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
	"time"
)

var (
	conbeeHost = flag.String("host", "127.0.0.1:80", "Conbee host addr")
	conbeeKey  = flag.String("key", "", "Conbee api key")
	addr       = flag.String("addr", ":8080", "Metrics http listen")
)

func main() {
	flag.Parse()

	logger, _ := zap.NewProduction()
	logger.Info("Starting deconz_exporter")

	if *conbeeKey == "" {
		logger.Fatal("A Conbee API key is required")
		return
	}

	prometheus.MustRegister(NewDeconzCollector("deconz", logger, sensors.New(*conbeeHost, *conbeeKey)))

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
            <head><title>Deconz Exporter</title></head>
            <body>
            <h1>Deconz Exporter</h1>
            <p><a href="/metrics">Metrics</a></p>
            </body>
            </html>`))
	})
	srv := &http.Server{
		Addr:         *addr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Info("Listening on", zap.Stringp("addr", addr))
	logger.Fatal("failed to start server", zap.Error(srv.ListenAndServe()))
}
