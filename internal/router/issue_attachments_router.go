package router

import (
	"IssueForge/internal/handler"
	"IssueForge/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func registerIssueAttachmentsRoutes(r *mux.Router, issueAttachmentsHandler *handler.IssueAttachmentsHandler, authMiddleware *middleware.AuthMiddleware) {
	r.Handle("/api/issues/{issueID}/attachments",
		authMiddleware.Authenticate(http.HandlerFunc(issueAttachmentsHandler.CreateAttachment)),
	).Methods("POST")

	r.Handle("/api/attachments/{attachmentID}",
		authMiddleware.Authenticate(http.HandlerFunc(issueAttachmentsHandler.GetAttachmentByID)),
	).Methods("GET")

	r.Handle("/api/issues/{issueID}/attachments",
		authMiddleware.Authenticate(http.HandlerFunc(issueAttachmentsHandler.ListIssueAttachments)),
	).Methods("GET")

	r.Handle("/api/comments/{commentID}/attachments",
		authMiddleware.Authenticate(http.HandlerFunc(issueAttachmentsHandler.ListCommentAttachments)),
	).Methods("GET")

	r.Handle("/api/attachments/{attachmentID}",
		authMiddleware.Authenticate(http.HandlerFunc(issueAttachmentsHandler.SoftDeleteAttachments)),
	).Methods("DELETE")
}
