package router

import (
	"IssueForge/internal/handler"
	"IssueForge/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func registerWorkspaceMemberRoutes(r *mux.Router, workspaceMemberHandler *handler.WorkspaceMemberHandler, authMiddleware *middleware.AuthMiddleware, readRateLimit *middleware.RateLimitMiddleware) {
	r.Handle("/api/workspaces/{workspaceID}/members",
		authMiddleware.Authenticate(http.HandlerFunc(workspaceMemberHandler.AddWorkspaceMember)),
	).Methods("POST")

	r.Handle("/api/workspaces/{workspaceID}/members/{userID}",
		authMiddleware.Authenticate(
			readRateLimit.Limit(http.HandlerFunc(workspaceMemberHandler.GetWorkspaceMember)),
		),
	).Methods("GET")

	r.Handle("/api/workspaces",
		authMiddleware.Authenticate(
			readRateLimit.Limit(http.HandlerFunc(workspaceMemberHandler.ListUserWorkspaces)),
		),
	).Methods("GET")

	r.Handle("/api/workspaces/{workspaceID}/members",
		authMiddleware.Authenticate(
			readRateLimit.Limit(http.HandlerFunc(workspaceMemberHandler.ListWorkspaceMembers)),
		),
	).Methods("GET")

	r.Handle("/api/workspaces/{workspaceID}/members/{userID}",
		authMiddleware.Authenticate(http.HandlerFunc(workspaceMemberHandler.RemoveWorkspaceMember)),
	).Methods("DELETE")
}
