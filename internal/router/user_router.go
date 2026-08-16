package router

import (
	"IssueForge/internal/handler"
	"IssueForge/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func registerUserRoutes(r *mux.Router, userHandler *handler.UserHandler, authMiddleware *middleware.AuthMiddleware, strictRateLimit, authRateLimit, readRateLimit, deleteRateLimit *middleware.RateLimitMiddleware) {
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

	r.Handle("/api/users/{userID}",
		authMiddleware.Authenticate(
			readRateLimit.Limit(http.HandlerFunc(userHandler.GetUserByID)),
		),
	).Methods("GET")

	r.Handle("/api/users/{username}",
		authMiddleware.Authenticate(
			readRateLimit.Limit(http.HandlerFunc(userHandler.GetUserByUsername)),
		),
	).Methods("GET")

	r.Handle("/api/users/search",
		authMiddleware.Authenticate(
			readRateLimit.Limit(http.HandlerFunc(userHandler.SearchUserByUsername)),
		),
	).Methods("GET")

	r.Handle("/api/users/{userID}",
		authMiddleware.Authenticate(
			deleteRateLimit.Limit(http.HandlerFunc(userHandler.DeleteUser)),
		),
	).Methods("DELETE")
}
