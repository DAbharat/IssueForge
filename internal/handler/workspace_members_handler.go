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

type WorkspaceMemberService interface {
	AddWorkspaceMember(ctx context.Context, adminID, workspaceID int64, req dto.AddWorkspaceMemberRequest) (dto.WorkspaceMemberResponse, error)
	GetWorkspaceMember(ctx context.Context, workspaceID, requesterID, targetUserID int64) (dto.WorkspaceMemberSummary, error)
	ListUserWorkspaces(ctx context.Context, userID int64, search string) ([]dto.WorkspaceSummary, error)
	ListWorkspaceMembers(ctx context.Context, workspaceID, userID int64) ([]dto.WorkspaceMemberDetails, error)
	RemoveWorkspaceMember(ctx context.Context, workspaceID, adminID, userID int64) (dto.RemoveWorkspaceMemberResponse, error)
}

type WorkspaceMemberHandler struct {
	workspaceMemberService WorkspaceMemberService
}

func NewWorkspaceMemberHandler(service WorkspaceMemberService) *WorkspaceMemberHandler {
	return &WorkspaceMemberHandler{
		workspaceMemberService: service,
	}
}

func (h *WorkspaceMemberHandler) AddWorkspaceMember(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	var req dto.AddWorkspaceMemberRequest

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

	adminID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)

	workspaceID, err := strconv.ParseInt(vars["workspaceID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	member, err := h.workspaceMemberService.AddWorkspaceMember(r.Context(), adminID, workspaceID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, repository.ErrWorkspaceMemberAlreadyExists):
			httpx.RespondWithError(w, http.StatusConflict, err.Error())
		case errors.Is(err, repository.ErrWorkspaceNotFound),
			errors.Is(err, repository.ErrUserNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrInvalidRole),
			errors.Is(err, service.ErrInvalidWorkspaceID),
			errors.Is(err, service.ErrInvalidUserID):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		default:
			log.Printf("add workspace member fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusCreated, member)
}

func (h *WorkspaceMemberHandler) GetWorkspaceMember(w http.ResponseWriter, r *http.Request) {
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

	userID, err := strconv.ParseInt(vars["userID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidUserID.Error())
		return
	}

	member, err := h.workspaceMemberService.GetWorkspaceMember(r.Context(), workspaceID, requesterID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, repository.ErrWorkspaceMemberNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("get workspace member fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, member)
}

func (h *WorkspaceMemberHandler) ListUserWorkspaces(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	search := r.URL.Query().Get("search")

	member, err := h.workspaceMemberService.ListUserWorkspaces(r.Context(), userID, search)
	if err != nil {
		log.Printf("list user workspaces fail: %v", err)
		httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, member)
}

func (h *WorkspaceMemberHandler) ListWorkspaceMembers(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserFromContext(r.Context())
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

	member, err := h.workspaceMemberService.ListWorkspaceMembers(r.Context(), workspaceID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		default:
			log.Printf("list workspace members fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	httpx.RespondWithJSON(w, http.StatusOK, member)
}

func (h *WorkspaceMemberHandler) RemoveWorkspaceMember(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.GetUserFromContext(r.Context())
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

	targetUserID, err := strconv.ParseInt(vars["userID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidUserID.Error())
		return
	}

	member, err := h.workspaceMemberService.RemoveWorkspaceMember(r.Context(), workspaceID, adminID, targetUserID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, repository.ErrWorkspaceMemberNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("remove workspace member fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, member)
}
