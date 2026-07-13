package router

import (
	"IssueForge/internal/handler"
	"IssueForge/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func registerWorkspaceRoutes(r *mux.Router, workspaceHandler *handler.WorkspaceHandler, authMiddleware *middleware.AuthMiddleware) {
	r.Handle("/api/workspaces",
		authMiddleware.Authenticate(http.HandlerFunc(workspaceHandler.CreateWorkspace)),
	).Methods("POST")

	r.Handle("/api/workspaces/{workspaceID}",
		authMiddleware.Authenticate(http.HandlerFunc(workspaceHandler.GetWorkspaceByID)),
	).Methods("GET")
}
