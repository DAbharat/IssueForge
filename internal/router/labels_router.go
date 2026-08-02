package router

import (
	"IssueForge/internal/handler"
	"IssueForge/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func registerLabelsRoutes(r *mux.Router, labelsHandler *handler.LabelsHandler, authMiddleware *middleware.AuthMiddleware) {
	r.Handle("/api/projects/{projectID}/labels",
		authMiddleware.Authenticate(http.HandlerFunc(labelsHandler.CreateLabel)),
	).Methods("POST")

	r.Handle("/api/labels/{labelID}",
		authMiddleware.Authenticate(http.HandlerFunc(labelsHandler.GetLabelByID)),
	).Methods("GET")

	r.Handle("/api/projects/{projectID}/labels",
		authMiddleware.Authenticate(http.HandlerFunc(labelsHandler.ListProjectLabels)),
	).Methods("GET")

	r.Handle("/api/labels/{labelID}",
		authMiddleware.Authenticate(http.HandlerFunc(labelsHandler.UpdateLabel)),
	).Methods("PATCH")

	r.Handle("/api/labels/{labelID}",
		authMiddleware.Authenticate(http.HandlerFunc(labelsHandler.DeleteLabel)),
	).Methods("DELETE")

	r.Handle("/api/issues/{issueID}/labels",
		authMiddleware.Authenticate(http.HandlerFunc(labelsHandler.AttachLabelsToIssue)),
	).Methods("POST")

	r.Handle("/api/issues/{issueID}/labels/{labelID}",
		authMiddleware.Authenticate(http.HandlerFunc(labelsHandler.RemoveLabelFromIssue)),
	).Methods("DELETE")

	r.Handle("/api/issues/{issueID}/labels",
		authMiddleware.Authenticate(http.HandlerFunc(labelsHandler.ListIssueLabels)),
	).Methods("GET")
}
