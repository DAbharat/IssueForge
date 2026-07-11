package handler

import (
	"IssueForge/internal/auth"
	"IssueForge/internal/dto"
	"IssueForge/internal/httpx"
	"IssueForge/internal/middleware"
	"IssueForge/internal/repository"
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type ProjectMemberService interface {
	AddMemberToProject(ctx context.Context, req dto.AddProjectMemberRequest) (dto.ProjectMemberResponse, error)
	ListProjectMembers(ctx context.Context, projectID, userID int64) ([]dto.ProjectMemberSummary, error)
	SafeAddMemberToProject(ctx context.Context, req dto.AddProjectMemberRequest, leadID int64) (dto.ProjectMemberResponse, error)
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
	leadID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)

	userID, err := strconv.ParseInt(vars["userID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	projectID, err := strconv.ParseInt(vars["projectID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	req := dto.AddProjectMemberRequest{
		ProjectID: projectID,
		UserID:    userID,
	}

	member, err := h.projectMemberService.SafeAddMemberToProject(r.Context(), req, leadID)
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
