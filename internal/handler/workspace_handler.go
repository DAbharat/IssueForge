package handler

import (
	"IssueForge/internal/dto"
	"IssueForge/internal/httpx"
	"IssueForge/internal/middleware"
	"IssueForge/internal/repository"
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

type WorkspaceService interface {
	CreateWorkspace(ctx context.Context, creatorID int64, req dto.CreateWorkspaceRequest) (dto.CreateWorkspaceResponse, error)
	GetWorkspaceByID(ctx context.Context, workspaceID, userID int64) (dto.WorkspaceResponse, error)
	GetWorkspaceByName(ctx context.Context, name string) (dto.WorkspaceResponse, error)
}

type WorkspaceHandler struct {
	workspaceService WorkspaceService
}

func NewWorkspaceHandler(service WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{
		workspaceService: service,
	}
}

func (h *WorkspaceHandler) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	var req dto.CreateWorkspaceRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid or oversizeed request body")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		httpx.RespondWithError(w, http.StatusBadRequest, "request body must contain a single JSON object")
		return
	}

	creatorID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	workspace, err := h.workspaceService.CreateWorkspace(r.Context(), creatorID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidWorkspaceName):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, repository.ErrWorkspaceAlreadyExists):
			httpx.RespondWithError(w, http.StatusConflict, err.Error())
		default:
			log.Printf("create workspace fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusCreated, workspace)
}

func (h *WorkspaceHandler) GetWorkspaceByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	userID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	workspaceID, err := strconv.ParseInt(vars["workspaceID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid workspace id")
		return
	}

	workspace, err := h.workspaceService.GetWorkspaceByID(r.Context(), workspaceID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, repository.ErrWorkspaceNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("get workspace by id fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, workspace)
}

func (h *WorkspaceHandler) GetWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	workspaceName := vars["workspaceName"]
	if workspaceName == "" {
		httpx.RespondWithError(w, http.StatusBadRequest, "workspace name is required")
		return
	}

	workspace, err := h.workspaceService.GetWorkspaceByName(r.Context(), workspaceName)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidWorkspaceName):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, repository.ErrWorkspaceNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("get workspace by name fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, workspace)
}
