package router

import (
	"IssueForge/internal/handler"
	"IssueForge/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func registerProjectActivityRouter(r *mux.Router, projectActivityHandler *handler.ProjectActivityHandler, authMiddleware *middleware.AuthMiddleware) {
	r.Handle("/api/projects/{projectID}/activities",
		authMiddleware.Authenticate(http.HandlerFunc(projectActivityHandler.ListProjectActivities)),
	).Methods("GET")
}
