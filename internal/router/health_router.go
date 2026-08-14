package router

import (
	"IssueForge/internal/handler"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func registerHealthRoutes(r *mux.Router, reg prometheus.Gatherer) {
	r.HandleFunc("/api/health", handler.HealthCheckHandler).Methods("GET")
	r.Handle("/api/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{})).Methods("GET")
}
