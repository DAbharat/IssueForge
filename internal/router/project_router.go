package router

import (
	"IssueForge/internal/handler"
	"IssueForge/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func registerProjectRoutes(r *mux.Router, projectHandler *handler.ProjectHandler, authMiddleware *middleware.AuthMiddleware, readRateLimit, createRateLimit, patchRateLimit *middleware.RateLimitMiddleware) {
	r.Handle("/api/workspaces/{workspaceID}/projects",
		authMiddleware.Authenticate(
			createRateLimit.Limit(http.HandlerFunc(projectHandler.CreateProject)),
		),
	).Methods("POST")

	r.Handle("/api/projects/{projectID}",
		authMiddleware.Authenticate(
			readRateLimit.Limit(http.HandlerFunc(projectHandler.GetProjectByID)),
		),
	).Methods("GET")

	r.Handle("/api/projects/{projectID}/details",
		authMiddleware.Authenticate(
			patchRateLimit.Limit(http.HandlerFunc(projectHandler.UpdateProjectDetails)),
		),
	).Methods("PATCH")

	r.Handle("/api/projects/{projectID}/lead",
		authMiddleware.Authenticate(
			patchRateLimit.Limit(http.HandlerFunc(projectHandler.UpdateProjectLead)),
		),
	).Methods("PATCH")

	r.Handle("/api/projects/{projectID}",
		authMiddleware.Authenticate(http.HandlerFunc(projectHandler.DeleteProject)),
	).Methods("DELETE")

	r.Handle("/api/workspaces/{workspaceID}/projects",
		authMiddleware.Authenticate(
			readRateLimit.Limit(http.HandlerFunc(projectHandler.ListProjectByLead)),
		),
	).Methods("GET")
}
