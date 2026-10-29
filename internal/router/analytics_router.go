package router

import (
	"IssueForge/internal/handler"
	"IssueForge/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func registerAnalyticsRoutes(
	r *mux.Router,
	analyticsHandler *handler.AnalyticsHandler,
	authMiddleware *middleware.AuthMiddleware,
	readRateLimit *middleware.RateLimitMiddleware,
) {
	r.Handle("/api/workspaces/{workspaceID}/analytics",
		authMiddleware.Authenticate(
			readRateLimit.Limit(http.HandlerFunc(analyticsHandler.GetWorkspaceAnalytics)),
		),
	).Methods("GET")

	r.Handle("/api/projects/{projectID}/analytics",
		authMiddleware.Authenticate(
			readRateLimit.Limit(http.HandlerFunc(analyticsHandler.GetProjectAnalytics)),
		),
	).Methods("GET")
}
