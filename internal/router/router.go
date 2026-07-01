package router

import (
	"IssueForge/internal/handler"

	"github.com/gorilla/mux"
)

func New(userHandler *handler.UserHandler) *mux.Router {
	r := mux.NewRouter()

	registerHealthRoutes(r)
	registerUserRoutes(r, userHandler)

	return r
}
