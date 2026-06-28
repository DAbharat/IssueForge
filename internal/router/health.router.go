package router

import (
	"IssueForge/internal/handler"

	"github.com/gorilla/mux"
)

func New() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/api/health", handler.HealthCheckHandler).Methods("GET")
	return r
}
