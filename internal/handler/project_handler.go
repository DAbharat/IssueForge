package handler

import (
	"IssueForge/internal/dto"
	"IssueForge/internal/middleware"
	"IssueForge/internal/service"
	"encoding/json"
	"errors"
	"net/http"
)

type ProjectHandler struct {
	projectService *service.ProjectService
}

func NewProjectHandler(service *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		projectService: service,
	}
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	var req dto.CreateProjectRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid or oversized request body")
		return
	}

	ownerID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	project, err := h.projectService.CreateProject(r.Context(), ownerID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidProjectName),
			errors.Is(err, service.ErrInvalidDescription):
			respondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrProjectNameTaken):
			respondWithError(w, http.StatusConflict, err.Error())
		default:
			respondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(project); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	projects, err := h.projectService.ListProjects(r.Context(), ownerID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(projects); err != nil {
		return
	}
}
