package router

import (
	"IssueForge/internal/handler"
	"IssueForge/internal/middleware"

	"github.com/gorilla/mux"
)

func New(userHandler *handler.UserHandler, projectHandler *handler.ProjectHandler, authMiddleware *middleware.AuthMiddleware) *mux.Router {
	r := mux.NewRouter()

	registerHealthRoutes(r)
	registerUserRoutes(r, userHandler, authMiddleware)
	registerProjectRoutes(r, projectHandler, authMiddleware)

	return r
}
