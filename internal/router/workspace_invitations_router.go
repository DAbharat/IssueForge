package router

import (
	"IssueForge/internal/handler"
	"IssueForge/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func registerWorkspaceInvitationsRoutes(r *mux.Router, workspaceInvitationsHandler *handler.WorkspaceInvitationHandler, createRateLimit, readRateLimit *middleware.RateLimitMiddleware, authMiddleware *middleware.AuthMiddleware) {
	r.Handle("/api/workspaces/{workspaceID}/invitations",
		authMiddleware.Authenticate(
			createRateLimit.Limit(http.HandlerFunc(workspaceInvitationsHandler.CreateWorkspaceInvitation)),
		),
	).Methods("POST")

	r.Handle("/api/invitations/{invitationID}",
		authMiddleware.Authenticate(
			readRateLimit.Limit(http.HandlerFunc(workspaceInvitationsHandler.GetWorkspaceInvitation)),
		),
	).Methods("GET")

	r.Handle("/api/invitations/pending",
		authMiddleware.Authenticate(
			readRateLimit.Limit(http.HandlerFunc(workspaceInvitationsHandler.ListPendingWorkspaceInvitations)),
		),
	).Methods("GET")

	r.Handle("/api/workspaces/{workspaceID}/invitations/pending",
		authMiddleware.Authenticate(
			readRateLimit.Limit(http.HandlerFunc(workspaceInvitationsHandler.ListPendingWorkspaceInvitationsForWorkspace)),
		),
	).Methods("GET")

	r.Handle("/api/invitations/{invitationID}/accept",
		authMiddleware.Authenticate(
			createRateLimit.Limit(http.HandlerFunc(workspaceInvitationsHandler.AcceptInvitation)),
		),
	).Methods("POST")

	r.Handle("/api/invitations/{invitationID}/decline",
		authMiddleware.Authenticate(
			createRateLimit.Limit(http.HandlerFunc(workspaceInvitationsHandler.DeclineInvitation)),
		),
	).Methods("POST")

	r.Handle("/api/invitations/{invitationID}/decline",
		authMiddleware.Authenticate(
			createRateLimit.Limit(http.HandlerFunc(workspaceInvitationsHandler.CancelInvitation)),
		),
	).Methods("POST")
}
