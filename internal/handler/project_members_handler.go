package handler

import (
	"IssueForge/internal/auth"
	"IssueForge/internal/dto"
	"IssueForge/internal/httpx"
	"IssueForge/internal/middleware"
	"IssueForge/internal/repository"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type ProjectMemberService interface {
	AddMemberToProject(ctx context.Context, projectID, leadID int64, req dto.AddProjectMemberRequest) (dto.ProjectMemberResponse, error)
	ListProjectMembers(ctx context.Context, projectID, userID int64) ([]dto.ProjectMemberSummary, error)
	SafeAddMemberToProject(ctx context.Context, req dto.AddProjectMemberRequest, projectID, leadID int64) (dto.ProjectMemberResponse, error)
}

type ProjectMemberHandler struct {
	projectMemberService ProjectMemberService
}

func NewProjectMemberHandler(service ProjectMemberService) *ProjectMemberHandler {
	return &ProjectMemberHandler{
		projectMemberService: service,
	}
}

func (h *ProjectMemberHandler) SafeAddMemberToProject(w http.ResponseWriter, r *http.Request) {
	var req dto.AddProjectMemberRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid or oversized request body")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		httpx.RespondWithError(w, http.StatusBadRequest, "request body must contain a single JSON object")
		return
	}

	leadID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)

	projectID, err := strconv.ParseInt(vars["projectID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	member, err := h.projectMemberService.SafeAddMemberToProject(r.Context(), req, projectID, leadID)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrForbidden):
			httpx.RespondWithError(w, http.StatusUnauthorized, err.Error())
		case errors.Is(err, repository.ErrProjectMemberValidationFailed):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, repository.ErrProjectMemberAlreadyExists):
			httpx.RespondWithError(w, http.StatusConflict, err.Error())
		case errors.Is(err, repository.ErrProjectNotFound),
			errors.Is(err, repository.ErrUserNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("safe add project member fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusCreated, member)
}

func (h *ProjectMemberHandler) ListProjectMembers(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)

	log.Printf("vars = %+v", vars)
	projectID, err := strconv.ParseInt(vars["projectID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	members, err := h.projectMemberService.ListProjectMembers(r.Context(), projectID, userID)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		default:
			log.Printf("list project by members fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, members)
}
