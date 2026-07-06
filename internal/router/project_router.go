package router

import (
	"IssueForge/internal/handler"
	"IssueForge/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func registerProjectRoutes(r *mux.Router, projectHandler *handler.ProjectHandler, authMiddleware *middleware.AuthMiddleware) {
	r.Handle("/api/projects",
		authMiddleware.Authenticate(http.HandlerFunc(projectHandler.Create)),
	).Methods("POST")

	r.Handle("/api/projects",
		authMiddleware.Authenticate(http.HandlerFunc(projectHandler.List)),
	).Methods("GET")
}
