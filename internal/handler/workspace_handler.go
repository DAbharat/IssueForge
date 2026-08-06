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
	UpdateWorkspaceName(ctx context.Context, requesterID, workspaceID int64, req dto.UpdateWorkspaceRequest) (dto.WorkspaceResponse, error)
	DeleteWorkspace(ctx context.Context, requesterID, workspaceID int64) (int64, error)
	RestoreDeletedWorkspace(ctx context.Context, requesterID, workspaceID int64) (dto.WorkspaceResponse, error)
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

func (h *WorkspaceHandler) UpdateWorkspaceName(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	var req dto.UpdateWorkspaceRequest

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

	workspaceID, err := strconv.ParseInt(vars["workspaceID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidWorkspaceID.Error())
		return
	}

	workspace, err := h.workspaceService.UpdateWorkspaceName(r.Context(), requesterID, workspaceID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidWorkspaceName):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrWorkspaceNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrWorkspaceNameTaken):
			httpx.RespondWithError(w, http.StatusConflict, err.Error())
		default:
			log.Printf("update workspace name fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, workspace)
}

func (h *WorkspaceHandler) DeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)

	workspaceID, err := strconv.ParseInt(vars["workspaceID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidWorkspaceID.Error())
		return
	}

	id, err := h.workspaceService.DeleteWorkspace(r.Context(), requesterID, workspaceID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrWorkspaceNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, id)
}

func (h *WorkspaceHandler) RestoreDeletedWorkspace(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)

	workspaceID, err := strconv.ParseInt(vars["workspaceID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidWorkspaceID.Error())
		return
	}

	workspace, err := h.workspaceService.RestoreDeletedWorkspace(r.Context(), requesterID, workspaceID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrWorkspaceNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("restore deleted workspace fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, workspace)
}
