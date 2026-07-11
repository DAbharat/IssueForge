package handler

import (
	"IssueForge/internal/auth"
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

type ProjectService interface {
	CreateProject(ctx context.Context, leadID int64, req dto.CreateProjectRequest) (dto.CreateProjectResponse, error)
	ListProjectByLead(ctx context.Context, leadID int64) ([]dto.ProjectResponse, error)
	ListProjectsByWorkspace(ctx context.Context, workspaceID, userID int64) ([]dto.ProjectResponse, error)
}

type ProjectHandler struct {
	projectService ProjectService
}

func NewProjectHandler(service ProjectService) *ProjectHandler {
	return &ProjectHandler{
		projectService: service,
	}
}

func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	var req dto.CreateProjectRequest

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

	project, err := h.projectService.CreateProject(r.Context(), leadID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidProjectName),
			errors.Is(err, service.ErrInvalidDescription):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrProjectNameTaken):
			httpx.RespondWithError(w, http.StatusConflict, err.Error())
		default:
			log.Printf("create project fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusCreated, project)
}

func (h *ProjectHandler) ListProjectByLead(w http.ResponseWriter, r *http.Request) {
	leadID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	projects, err := h.projectService.ListProjectByLead(r.Context(), leadID)
	if err != nil {
		log.Printf("list project by lead fail: %v", err)
		httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, projects)
}

func (h *ProjectHandler) ListProjectsByWorkspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	workspaceID, err := strconv.ParseInt(vars["workspaceID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid workspace id")
		return
	}

	userID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	projects, err := h.projectService.ListProjectsByWorkspace(r.Context(), workspaceID, userID)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		default:
			log.Printf("list project by workspace fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, projects)
}
