package router

import (
	"IssueForge/internal/handler"
	"IssueForge/internal/middleware"

	"github.com/gorilla/mux"
)

func New(userHandler *handler.UserHandler, authMiddleware *middleware.AuthMiddleware) *mux.Router {
	r := mux.NewRouter()

	registerHealthRoutes(r)
	registerUserRoutes(r, userHandler, authMiddleware)

	return r
}
