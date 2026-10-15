package router

import (
	"IssueForge/internal/handler"
	"IssueForge/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func registerArchiveRoutes(r *mux.Router, archiveHandler *handler.ArchiveHandler, authMiddleware *middleware.AuthMiddleware, readRateLimit *middleware.RateLimitMiddleware) {
	r.Handle("/api/archive/counts",
		authMiddleware.Authenticate(
			readRateLimit.Limit(http.HandlerFunc(archiveHandler.GetArchivedCounts)),
		),
	).Methods("GET")

	r.Handle("/api/archive/workspaces",
		authMiddleware.Authenticate(
			readRateLimit.Limit(http.HandlerFunc(archiveHandler.ListArchivedWorkspaces)),
		),
	).Methods("GET")

	r.Handle("/api/archive/projects",
		authMiddleware.Authenticate(
			readRateLimit.Limit(http.HandlerFunc(archiveHandler.ListArchivedProjects)),
		),
	).Methods("GET")

	r.Handle("/api/archive/issues",
		authMiddleware.Authenticate(
			readRateLimit.Limit(http.HandlerFunc(archiveHandler.ListArchivedIssues)),
		),
	).Methods("GET")

	r.Handle("/api/archive/attachments",
		authMiddleware.Authenticate(
			readRateLimit.Limit(http.HandlerFunc(archiveHandler.ListArchivedAttachments)),
		),
	).Methods("GET")
}
