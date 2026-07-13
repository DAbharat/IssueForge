package router

import (
	"IssueForge/internal/handler"
	"IssueForge/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func registerProjectMembersRoutes(r *mux.Router, projectMemberHandler *handler.ProjectMemberHandler, authMiddleware *middleware.AuthMiddleware) {
	r.Handle("/api/projects/{projectID}/members",
		authMiddleware.Authenticate(http.HandlerFunc(projectMemberHandler.SafeAddMemberToProject)),
	).Methods("POST")

	r.Handle("/api/projects/{projectID}/members",
		authMiddleware.Authenticate(http.HandlerFunc(projectMemberHandler.ListProjectMembers)),
	).Methods("GET")
}
