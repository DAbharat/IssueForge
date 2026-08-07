package router

import (
	"IssueForge/internal/handler"
	"IssueForge/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func registerIssueAttachmentsRoutes(r *mux.Router, issueAttachmentsHandler *handler.IssueAttachmentsHandler, authMiddleware *middleware.AuthMiddleware, attachmentRateLimit, readRateLimit *middleware.RateLimitMiddleware) {
	r.Handle("/api/issues/{issueID}/attachments",
		authMiddleware.Authenticate(
			attachmentRateLimit.Limit(http.HandlerFunc(issueAttachmentsHandler.CreateAttachment)),
		),
	).Methods("POST")

	r.Handle("/api/attachments/{attachmentID}",
		authMiddleware.Authenticate(
			readRateLimit.Limit(http.HandlerFunc(issueAttachmentsHandler.GetAttachmentByID)),
		),
	).Methods("GET")

	r.Handle("/api/issues/{issueID}/attachments",
		authMiddleware.Authenticate(
			readRateLimit.Limit(http.HandlerFunc(issueAttachmentsHandler.ListIssueAttachments)),
		),
	).Methods("GET")

	r.Handle("/api/comments/{commentID}/attachments",
		authMiddleware.Authenticate(
			readRateLimit.Limit(http.HandlerFunc(issueAttachmentsHandler.ListCommentAttachments)),
		),
	).Methods("GET")

	r.Handle("/api/attachments/{attachmentID}",
		authMiddleware.Authenticate(http.HandlerFunc(issueAttachmentsHandler.SoftDeleteAttachments)),
	).Methods("DELETE")
}
