package router

import (
	"IssueForge/internal/handler"
	"IssueForge/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func registerProjectRoutes(r *mux.Router, projectHandler *handler.ProjectHandler, authMiddleware *middleware.AuthMiddleware) {
	r.Handle("/api/workspaces/{workspaceID}/projects",
		authMiddleware.Authenticate(http.HandlerFunc(projectHandler.CreateProject)),
	).Methods("POST")

	r.Handle("/api/projects/{projectID}",
		authMiddleware.Authenticate(http.HandlerFunc(projectHandler.GetProjectByID)),
	).Methods("GET")

	r.Handle("/api/projects/{projectID}/details",
		authMiddleware.Authenticate(http.HandlerFunc(projectHandler.UpdateProjectDetails)),
	).Methods("PATCH")

	r.Handle("/api/projects/{projectID}/lead",
		authMiddleware.Authenticate(http.HandlerFunc(projectHandler.UpdateProjectLead)),
	).Methods("PATCH")

	r.Handle("/api/projects/{projectID}",
		authMiddleware.Authenticate(http.HandlerFunc(projectHandler.DeleteProject)),
	).Methods("DELETE")

	r.Handle("/api/workspaces/{workspaceID}/projects",
		authMiddleware.Authenticate(http.HandlerFunc(projectHandler.ListProjectByLead)),
	).Methods("GET")
}
