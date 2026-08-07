package router

import (
	"IssueForge/internal/handler"
	"IssueForge/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func registerProjectMembersRoutes(r *mux.Router, projectMemberHandler *handler.ProjectMemberHandler, authMiddleware *middleware.AuthMiddleware, readRateLimit *middleware.RateLimitMiddleware) {
	r.Handle("/api/projects/{projectID}/members",
		authMiddleware.Authenticate(http.HandlerFunc(projectMemberHandler.SafeAddMemberToProject)),
	).Methods("POST")

	r.Handle("/api/projects/{projectID}/members",
		authMiddleware.Authenticate(
			readRateLimit.Limit(http.HandlerFunc(projectMemberHandler.ListProjectMembers)),
		),
	).Methods("GET")
}
