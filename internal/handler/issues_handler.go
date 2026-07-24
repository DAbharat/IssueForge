package handler

import (
	"IssueForge/internal/dto"
	"IssueForge/internal/httpx"
	"IssueForge/internal/middleware"
	"IssueForge/internal/service"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type IssueService interface {
	CreateIssue(ctx context.Context, creatorID int64, req dto.CreateIssueRequest) (dto.CreateIssueResponse, error)
	GetIssueByID(ctx context.Context, requesterID, issueID int64) (dto.IssueResponse, error)
	ListProjectIssues(ctx context.Context, requesterID, projectID int64, req dto.ListProjectIssuesRequest) ([]dto.IssueSummary, error)
	UpdateIssueDetails(ctx context.Context, requesterID, issueID int64, req dto.UpdateIssueDetailsRequest) (dto.IssueResponse, error)
	UpdateIssueStatus(ctx context.Context, requesterID, issueID int64, req dto.UpdateIssueStatusRequest) (dto.IssueResponse, error)
	UpdateIssueAssignee(ctx context.Context, requesterID, issueID int64, req dto.UpdateIssueAssigneeRequest) (dto.IssueResponse, error)
	UpdateIssuePriority(ctx context.Context, requesterID, issueID int64, req dto.UpdateIssuePriority) (dto.IssueResponse, error)
	ListAssignedIssues(ctx context.Context, requesterID, assignedTo int64) ([]dto.UserIssueSummary, error)
	ListCreatedIssues(ctx context.Context, requesterID, createdBy int64) ([]dto.UserIssueSummary, error)
	DeleteIssue(ctx context.Context, requesterID, issueID int64) (int64, error)
	RestoreDeletedIssue(ctx context.Context, requesterID, issueID int64) (int64, error)
}

type IssueHandler struct {
	issueService IssueService
}

func NewIssueHandler(service IssueService) *IssueHandler {
	return &IssueHandler{
		issueService: service,
	}
}

func (h *IssueHandler) CreateIssue(w http.ResponseWriter, r *http.Request) {
	creatorID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	var req dto.CreateIssueRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid or oversized request body")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		httpx.RespondWithError(w, http.StatusBadRequest, "request body must contain a single json object")
		return
	}

	vars := mux.Vars(r)

	projectID, err := strconv.ParseInt(vars["projectID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidPassword.Error())
		return
	}

	req.ProjectID = projectID

	issue, err := h.issueService.CreateIssue(r.Context(), creatorID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidProjectID),
			errors.Is(err, service.ErrInvalidAssignee),
			errors.Is(err, service.ErrInvalidTitle),
			errors.Is(err, service.ErrInvalidDescription),
			errors.Is(err, service.ErrInvalidStatus),
			errors.Is(err, service.ErrInvalidPriority):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrProjectNotFound),
			errors.Is(err, service.ErrUserNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("create issue fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusCreated, issue)
}

func (h *IssueHandler) GetIssueByID(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)

	issueID, err := strconv.ParseInt(vars["issueID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidIssueID.Error())
		return
	}

	issue, err := h.issueService.GetIssueByID(r.Context(), requesterID, issueID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidIssueID):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrIssueNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("get issue by id fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, issue)
}

func (h *IssueHandler) ListProjectIssues(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	req := dto.ListProjectIssuesRequest{
		Limit:  20,
		Offset: 0,
	}

	if value := r.URL.Query().Get("status"); value != "" {
		req.Status = &value
	}
	if value := r.URL.Query().Get("priority"); value != "" {
		req.Priority = &value
	}
	if value := r.URL.Query().Get("search"); value != "" {
		req.Search = &value
	}

	if value := r.URL.Query().Get("assigend_to"); value != "" {
		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidAssignee.Error())
			return
		}
		req.AssignedTo = &id
	}

	if value := r.URL.Query().Get("limit"); value != "" {
		limit, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidLimit.Error())
			return
		}
		req.Limit = int32(limit)
	}

	if value := r.URL.Query().Get("offset"); value != "" {
		offset, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidOffset.Error())
			return
		}
		req.Offset = int32(offset)
	}

	vars := mux.Vars(r)

	projectID, err := strconv.ParseInt(vars["projectID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidProjectID.Error())
		return
	}

	issue, err := h.issueService.ListProjectIssues(r.Context(), requesterID, projectID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrProjectNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrInvalidProjectID):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		default:
			log.Printf("list project issues fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, issue)
}

func (h *IssueHandler) UpdateIssueDetails(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.UpdateIssueDetailsRequest

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid or oversized request body")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		httpx.RespondWithError(w, http.StatusBadRequest, "request body must contain a single json object")
		return
	}

	vars := mux.Vars(r)

	issueID, err := strconv.ParseInt(vars["issueID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidIssueID.Error())
		return
	}

	issue, err := h.issueService.UpdateIssueDetails(r.Context(), requesterID, issueID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidIssueID):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrIssueNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrInvalidTitle),
			errors.Is(err, service.ErrInvalidDescription):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		default:
			log.Printf("update issue details fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, issue)
}

func (h *IssueHandler) UpdateIssueStatus(w http.ResponseWriter, r *http.Request) {
	log.Println("===== UpdateIssueStatus handler called =====")
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.UpdateIssueStatusRequest
	log.Printf("req body: %v", r.Body)

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		log.Printf("decode error: %T: %v", err, err)
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid or oversized request body")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		httpx.RespondWithError(w, http.StatusBadRequest, "request body must contain a single json object")
		return
	}

	vars := mux.Vars(r)

	issueID, err := strconv.ParseInt(vars["issueID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidIssueID.Error())
		return
	}

	issue, err := h.issueService.UpdateIssueStatus(r.Context(), requesterID, issueID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidIssueID):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrIssueNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrInvalidStatus):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		default:
			log.Printf("update issue status fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, issue)
}

func (h *IssueHandler) UpdateIssueAssignee(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	var req dto.UpdateIssueAssigneeRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid or oversized request body")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		httpx.RespondWithError(w, http.StatusBadRequest, "request body must contain a single json object")
		return
	}

	vars := mux.Vars(r)

	issueID, err := strconv.ParseInt(vars["issueID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidIssueID.Error())
		return
	}

	issue, err := h.issueService.UpdateIssueAssignee(r.Context(), requesterID, issueID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidIssueID),
			errors.Is(err, service.ErrInvalidAssignee):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrIssueNotFound),
			errors.Is(err, service.ErrUserNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("update issue assignee fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, issue)
}

func (h *IssueHandler) UpdateIssuePriority(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusForbidden, "unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	var req dto.UpdateIssuePriority

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid or oversized request body")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		httpx.RespondWithError(w, http.StatusBadRequest, "request body must contain a single json object")
		return
	}

	vars := mux.Vars(r)

	issueID, err := strconv.ParseInt(vars["issueID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidIssueID.Error())
		return
	}

	issue, err := h.issueService.UpdateIssuePriority(r.Context(), requesterID, issueID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrIssueNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrInvalidIssueID),
			errors.Is(err, service.ErrInvalidPriority):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		default:
			log.Printf("update issue priority fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, issue)
}

func (h *IssueHandler) ListAssignedIssues(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	assignedTo := requesterID

	issues, err := h.issueService.ListAssignedIssues(r.Context(), requesterID, assignedTo)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidAssignee):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		default:
			log.Printf("list assigned issues fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, issues)
}

func (h *IssueHandler) ListCreatedIssues(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	createdBy := requesterID

	issues, err := h.issueService.ListCreatedIssues(r.Context(), requesterID, createdBy)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidUserID):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		default:
			log.Printf("list created issues fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, issues)
}

func (h *IssueHandler) DeleteIssue(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)

	issueID, err := strconv.ParseInt(vars["issueID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidIssueID.Error())
		return
	}

	issue, err := h.issueService.DeleteIssue(r.Context(), requesterID, issueID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidIssueID):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrIssueNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("delete issue fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, map[string]int64{
		"deleted_issue_id": issue,
	})
}

func (h *IssueHandler) RestoreDeletedIssue(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)

	issueID, err := strconv.ParseInt(vars["issueID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidIssueID.Error())
		return
	}

	issue, err := h.issueService.RestoreDeletedIssue(r.Context(), requesterID, issueID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidIssueID):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrIssueNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("restore deleted issue fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, issue)
}
