package router

import (
	"IssueForge/internal/handler"
	"IssueForge/internal/middleware"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

func New(userHandler *handler.UserHandler,
	projectHandler *handler.ProjectHandler,
	projectMemberHandler *handler.ProjectMemberHandler,
	workspaceHandler *handler.WorkspaceHandler,
	workspaceMemberHandler *handler.WorkspaceMemberHandler,
	issueHandler *handler.IssueHandler,
	commentHandler *handler.CommentHandler,
	issueActivityHandler *handler.IssueActivityHandler,
	issueAttachmentsHandler *handler.IssueAttachmentsHandler,
	labelsHandler *handler.LabelsHandler,
	authMiddleware *middleware.AuthMiddleware,
	reg *prometheus.Registry,
	strictRateLimit *middleware.RateLimitMiddleware,
	authRateLimit *middleware.RateLimitMiddleware,
	readRateLimit *middleware.RateLimitMiddleware,
	attachmentRateLimit *middleware.RateLimitMiddleware,
	writeRateLimit *middleware.RateLimitMiddleware,
	issueRateLimit *middleware.RateLimitMiddleware,
	deleteRateLimit *middleware.RateLimitMiddleware,
	createRateLimit *middleware.RateLimitMiddleware,
	patchRateLimit *middleware.RateLimitMiddleware,
) *mux.Router {

	r := mux.NewRouter()

	registerHealthRoutes(r, reg)
	registerUserRoutes(r, userHandler, authMiddleware, strictRateLimit, authRateLimit, readRateLimit, deleteRateLimit)
	registerProjectRoutes(r, projectHandler, authMiddleware, readRateLimit, createRateLimit, patchRateLimit)
	registerProjectMembersRoutes(r, projectMemberHandler, authMiddleware, readRateLimit)
	registerWorkspaceRoutes(r, workspaceHandler, authMiddleware, readRateLimit, createRateLimit, patchRateLimit)
	registerWorkspaceMemberRoutes(r, workspaceMemberHandler, authMiddleware, readRateLimit)
	registerIssueRoutes(r, issueHandler, authMiddleware, createRateLimit, readRateLimit, deleteRateLimit, patchRateLimit)
	registerCommentRoutes(r, commentHandler, authMiddleware)
	registerIssueActivityRouter(r, issueActivityHandler, authMiddleware)
	registerIssueAttachmentsRoutes(r, issueAttachmentsHandler, authMiddleware, attachmentRateLimit, readRateLimit)
	registerLabelsRoutes(r, labelsHandler, authMiddleware)

	metrics := middleware.NewMetricsMiddleware(reg)
	r.Use(metrics.Metrics)

	return r
}
