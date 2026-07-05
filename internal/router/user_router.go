package router

import (
	"IssueForge/internal/handler"
	"IssueForge/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func registerUserRoutes(r *mux.Router, userHandler *handler.UserHandler, authMiddleware *middleware.AuthMiddleware) {
	r.HandleFunc("/api/register", userHandler.Signup).Methods("POST")
	r.HandleFunc("/api/login", userHandler.Login).Methods("POST")
	r.Handle(
		"/api/me",
		authMiddleware.Authenticate(
			http.HandlerFunc(userHandler.Me),
		),
	).Methods("GET")
}
