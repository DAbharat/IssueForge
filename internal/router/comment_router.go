package router

import (
	"IssueForge/internal/handler"
	"IssueForge/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func registerCommentRoutes(r *mux.Router, commentHandler *handler.CommentHandler, authMiddleware *middleware.AuthMiddleware) {
	r.Handle("/api/issues/{issueID}/comments",
		authMiddleware.Authenticate(http.HandlerFunc(commentHandler.CreateComment)),
	).Methods("POST")

	r.Handle("/api/comments/{commentID}",
		authMiddleware.Authenticate(http.HandlerFunc(commentHandler.GetCommentByID)),
	).Methods("GET")

	r.Handle("/api/issues/{issueID}/comments",
		authMiddleware.Authenticate(http.HandlerFunc(commentHandler.ListIssueComments)),
	).Methods("GET")

	r.Handle("/api/comments/{commentID}",
		authMiddleware.Authenticate(http.HandlerFunc(commentHandler.UpdateComment)),
	).Methods("PATCH")

	r.Handle("/api/comments/{commentID}",
		authMiddleware.Authenticate(http.HandlerFunc(commentHandler.DeleteComment)),
	).Methods("DELETE")
}
