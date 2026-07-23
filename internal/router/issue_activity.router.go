package router

import (
	"IssueForge/internal/handler"
	"IssueForge/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func registerIssueActivityRouter(r *mux.Router, issueActivityHandler *handler.IssueActivityHandler, authMiddleware *middleware.AuthMiddleware) {
	r.Handle("/api/issues/{issueID}/activities",
		authMiddleware.Authenticate(http.HandlerFunc(issueActivityHandler.ListIssueActivities)),
	).Methods("GET")
}
