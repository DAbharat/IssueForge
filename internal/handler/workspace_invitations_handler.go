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

type WorkspaceInvitationsService interface {
	CreateWorkspaceInvitation(ctx context.Context, workspaceID int64, req dto.CreateWorkspaceInvitationRequest, requesterID int64) (dto.WorkspaceInvitationResponse, error)
	GetWorkspaceInvitation(ctx context.Context, requesterID, invitationID int64) (dto.WorkspaceInvitationResponse, error)
	ListPendingWorkspaceInvitations(ctx context.Context, requesterID int64) ([]dto.PendingWorkspaceInvitationResponse, error)
	ListPendingWorkspaceInvitationsForWorkspace(ctx context.Context, requesterID, workspaceID int64) ([]dto.WorkspacePendingInvitationResponse, error)
	AcceptInvitation(ctx context.Context, invitationID, requesterID int64) (dto.WorkspaceInvitationResponse, error)
	DeclineInvitation(ctx context.Context, invitationID, requesterID int64) (dto.WorkspaceInvitationResponse, error)
	CancelInvitation(ctx context.Context, invitationID, requesterID int64) (dto.WorkspaceInvitationResponse, error)
}

type WorkspaceInvitationHandler struct {
	workspaceInvitationService WorkspaceInvitationsService
}

func NewWorkspaceInvitationHandler(workspaceInvitationService WorkspaceInvitationsService) *WorkspaceInvitationHandler {
	return &WorkspaceInvitationHandler{
		workspaceInvitationService: workspaceInvitationService,
	}
}

func (h *WorkspaceInvitationHandler) CreateWorkspaceInvitation(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.CreateWorkspaceInvitationRequest

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

	workspaceID, err := strconv.ParseInt(vars["workspaceID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidWorkspaceID.Error())
		return
	}

	invitation, err := h.workspaceInvitationService.CreateWorkspaceInvitation(r.Context(), workspaceID, req, requesterID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrPendingInvitationExists):
			httpx.RespondWithError(w, http.StatusConflict, err.Error())
		default:
			log.Printf("create workspace invitation fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusCreated, invitation)
}

func (h *WorkspaceInvitationHandler) GetWorkspaceInvitation(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)

	invitationID, err := strconv.ParseInt(vars["invitationID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidInvitationID.Error())
		return
	}

	invitation, err := h.workspaceInvitationService.GetWorkspaceInvitation(r.Context(), requesterID, invitationID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvitationNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("get invitation fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, invitation)
}

func (h *WorkspaceInvitationHandler) ListPendingWorkspaceInvitations(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	invitations, err := h.workspaceInvitationService.ListPendingWorkspaceInvitations(r.Context(), requesterID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		default:
			log.Printf("list pending workspace invitations fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, invitations)
}

func (h *WorkspaceInvitationHandler) ListPendingWorkspaceInvitationsForWorkspace(w http.ResponseWriter, r *http.Request) {
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

	invitations, err := h.workspaceInvitationService.ListPendingWorkspaceInvitationsForWorkspace(r.Context(), requesterID, workspaceID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		default:
			log.Printf("list workspace invitations for workspace: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, invitations)
}

func (h *WorkspaceInvitationHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)

	invitationID, err := strconv.ParseInt(vars["invitationID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidInvitationID.Error())
		return
	}

	invitations, err := h.workspaceInvitationService.AcceptInvitation(r.Context(), invitationID, requesterID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvitationNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrWorkspaceMemberAlreadyExists):
			httpx.RespondWithError(w, http.StatusConflict, err.Error())
		default:
			log.Printf("accept invitation: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, invitations)
}

func (h *WorkspaceInvitationHandler) DeclineInvitation(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)

	invitationID, err := strconv.ParseInt(vars["invitationID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidInvitationID.Error())
		return
	}

	invitation, err := h.workspaceInvitationService.DeclineInvitation(r.Context(), invitationID, requesterID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvitationNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("decline invitation: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, invitation)
}

func (h *WorkspaceInvitationHandler) CancelInvitation(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)

	invitationID, err := strconv.ParseInt(vars["invitationID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidInvitationID.Error())
		return
	}

	invitation, err := h.workspaceInvitationService.CancelInvitation(r.Context(), invitationID, requesterID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvitationNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("cancel invitation: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, invitation)
}
