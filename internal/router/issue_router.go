package router

import (
	"IssueForge/internal/handler"
	"IssueForge/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func registerIssueRoutes(r *mux.Router, issueHandler *handler.IssueHandler, authMiddleware *middleware.AuthMiddleware) {
	r.Handle("/api/projects/{projectID}/issues",
		authMiddleware.Authenticate(http.HandlerFunc(issueHandler.CreateIssue)),
	).Methods("POST")

	r.Handle("/api/issues/{issueID}",
		authMiddleware.Authenticate(http.HandlerFunc(issueHandler.GetIssueByID)),
	).Methods("GET")

	r.Handle("/api/projects/{projectID}/issues",
		authMiddleware.Authenticate(http.HandlerFunc(issueHandler.ListProjectIssues)),
	).Methods("GET")

	r.Handle("/api/issues/{issueID}/details",
		authMiddleware.Authenticate(http.HandlerFunc(issueHandler.UpdateIssueDetails)),
	).Methods("PATCH")

	r.Handle("/api/issues/{issueID}/status",
		authMiddleware.Authenticate(http.HandlerFunc(issueHandler.UpdateIssueStatus)),
	).Methods("PATCH")

	r.Handle("/api/issues/{issueID}/assignee",
		authMiddleware.Authenticate(http.HandlerFunc(issueHandler.UpdateIssueAssignee)),
	).Methods("PATCH")

	r.Handle("/api/issues/{issueID}/priority",
		authMiddleware.Authenticate(http.HandlerFunc(issueHandler.UpdateIssuePriority)),
	).Methods("PATCH")

	r.Handle("/api/users/me/issues/assigned",
		authMiddleware.Authenticate(http.HandlerFunc(issueHandler.ListAssignedIssues)),
	).Methods("GET")

	r.Handle("/api/users/me/issues/created",
		authMiddleware.Authenticate(http.HandlerFunc(issueHandler.ListCreatedIssues)),
	).Methods("GET")

	r.Handle("/api/issues/{issueID}/due-date",
		authMiddleware.Authenticate(http.HandlerFunc(issueHandler.UpdateIssueDueDate)),
	).Methods("PATCH")

	r.Handle("/api/issues/{issueID}",
		authMiddleware.Authenticate(http.HandlerFunc(issueHandler.DeleteIssue)),
	).Methods("DELETE")

	r.Handle("/api/issues/{issueID}/restore",
		authMiddleware.Authenticate(http.HandlerFunc(issueHandler.RestoreDeletedIssue)),
	).Methods("PATCH")
}
