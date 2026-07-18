package router

import (
	"IssueForge/internal/handler"
	"IssueForge/internal/middleware"

	"github.com/gorilla/mux"
)

func New(userHandler *handler.UserHandler, projectHandler *handler.ProjectHandler, projectMemberHandler *handler.ProjectMemberHandler, workspaceHandler *handler.WorkspaceHandler, workspaceMemberHandler *handler.WorkspaceMemberHandler, issueHandler *handler.IssueHandler, authMiddleware *middleware.AuthMiddleware) *mux.Router {
	r := mux.NewRouter()

	registerHealthRoutes(r)
	registerUserRoutes(r, userHandler, authMiddleware)
	registerProjectRoutes(r, projectHandler, authMiddleware)
	registerProjectMembersRoutes(r, projectMemberHandler, authMiddleware)
	registerWorkspaceRoutes(r, workspaceHandler, authMiddleware)
	registerWorkspaceMemberRoutes(r, workspaceMemberHandler, authMiddleware)
	registerIssueRoutes(r, issueHandler, authMiddleware)

	return r
}
