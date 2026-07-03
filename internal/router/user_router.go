package router

import (
	"IssueForge/internal/handler"

	"github.com/gorilla/mux"
)

func registerUserRoutes(r *mux.Router, userHandler *handler.UserHandler) {
	r.HandleFunc("/api/register", userHandler.Signup).Methods("POST")
	r.HandleFunc("/api/login", userHandler.Login).Methods("POST")
}
