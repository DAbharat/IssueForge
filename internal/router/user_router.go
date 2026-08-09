package router

import (
	"IssueForge/internal/handler"
	"IssueForge/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func registerUserRoutes(r *mux.Router, userHandler *handler.UserHandler, authMiddleware *middleware.AuthMiddleware, strictRateLimit, authRateLimit, readRateLimit *middleware.RateLimitMiddleware) {
	r.Handle("/api/register",
		authRateLimit.Limit(http.HandlerFunc(userHandler.Signup)),
	).Methods("POST")

	r.Handle("/api/login",
		strictRateLimit.Limit(http.HandlerFunc(userHandler.Login)),
	).Methods("POST")

	r.Handle("/api/auth/refresh",
		authRateLimit.Limit(http.HandlerFunc(userHandler.RefreshAccessToken)),
	).Methods("POST")

	r.Handle(
		"/api/me",
		authMiddleware.Authenticate(
			readRateLimit.Limit(
				http.HandlerFunc(userHandler.Me),
			),
		),
	).Methods("GET")
}
