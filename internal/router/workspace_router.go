package router

import (
	"IssueForge/internal/handler"
	"IssueForge/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func registerWorkspaceRoutes(r *mux.Router, workspaceHandler *handler.WorkspaceHandler, authMiddleware *middleware.AuthMiddleware, readRateLimit, createRateLimit, patchRateLimit *middleware.RateLimitMiddleware) {
	r.Handle("/api/workspaces",
		authMiddleware.Authenticate(
			createRateLimit.Limit(http.HandlerFunc(workspaceHandler.CreateWorkspace)),
		),
	).Methods("POST")

	r.Handle("/api/workspaces/{workspaceID}",
		authMiddleware.Authenticate(
			readRateLimit.Limit(http.HandlerFunc(workspaceHandler.GetWorkspaceByID)),
		),
	).Methods("GET")

	r.Handle("/api/workspaces/{workspaceID}",
		authMiddleware.Authenticate(
			patchRateLimit.Limit(http.HandlerFunc(workspaceHandler.UpdateWorkspaceName)),
		),
	).Methods("PATCH")

	r.Handle("/api/workspaces/{workspaceID}",
		authMiddleware.Authenticate(http.HandlerFunc(workspaceHandler.DeleteWorkspace)),
	).Methods("DELETE")

	r.Handle("/api/workspace/{workspaceID}",
		authMiddleware.Authenticate(http.HandlerFunc(workspaceHandler.RestoreDeletedWorkspace)),
	).Methods("POST")
}
