package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/qingwave/gocorex/metrics"
)

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	state := metrics.NewHttpState(metrics.WithNamespace("demo"), metrics.WithRegistry(prometheus.NewRegistry()))

	http.Handle("/metrics", state.Handler())

	handler := state.WrapMetrics(http.DefaultServeMux)

	addr := "127.0.0.1:8080"
	log.Printf("start at: %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
