package handler

import (
	"IssueForge/internal/dto"
	"IssueForge/internal/httpx"
	"IssueForge/internal/middleware"
	"context"
	"log"
	"net/http"
)

type ArchiveService interface {
	GetArchivedCounts(ctx context.Context, userID int64) (dto.ArchiveCountsResponse, error)
	ListArchivedWorkspaces(ctx context.Context, userID int64, search *string) ([]dto.ArchivedWorkspaceItem, error)
	ListArchivedProjects(ctx context.Context, userID int64, search *string) ([]dto.ArchivedProjectItem, error)
	ListArchivedIssues(ctx context.Context, userID int64, search *string) ([]dto.ArchivedIssueItem, error)
	ListArchivedAttachments(ctx context.Context, userID int64, search *string) ([]dto.ArchivedAttachmentItem, error)
}

type ArchiveHandler struct {
	archiveService ArchiveService
}

func NewArchiveHandler(service ArchiveService) *ArchiveHandler {
	return &ArchiveHandler{archiveService: service}
}

func (h *ArchiveHandler) GetArchivedCounts(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	counts, err := h.archiveService.GetArchivedCounts(r.Context(), requesterID)
	if err != nil {
		log.Printf("get archived counts fail: %v", err)
		httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httpx.RespondWithJSON(w, http.StatusOK, counts)
}

func (h *ArchiveHandler) ListArchivedWorkspaces(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var search *string
	if s := r.URL.Query().Get("search"); s != "" {
		search = &s
	}

	workspaces, err := h.archiveService.ListArchivedWorkspaces(r.Context(), requesterID, search)
	if err != nil {
		log.Printf("list archived workspaces fail: %v", err)
		httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httpx.RespondWithJSON(w, http.StatusOK, workspaces)
}

func (h *ArchiveHandler) ListArchivedProjects(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var search *string
	if s := r.URL.Query().Get("search"); s != "" {
		search = &s
	}

	projects, err := h.archiveService.ListArchivedProjects(r.Context(), requesterID, search)
	if err != nil {
		log.Printf("list archived projects fail: %v", err)
		httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httpx.RespondWithJSON(w, http.StatusOK, projects)
}

func (h *ArchiveHandler) ListArchivedIssues(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var search *string
	if s := r.URL.Query().Get("search"); s != "" {
		search = &s
	}

	issues, err := h.archiveService.ListArchivedIssues(r.Context(), requesterID, search)
	if err != nil {
		log.Printf("list archived issues fail: %v", err)
		httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httpx.RespondWithJSON(w, http.StatusOK, issues)
}

func (h *ArchiveHandler) ListArchivedAttachments(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var search *string
	if s := r.URL.Query().Get("search"); s != "" {
		search = &s
	}

	attachments, err := h.archiveService.ListArchivedAttachments(r.Context(), requesterID, search)
	if err != nil {
		log.Printf("list archived attachments fail: %v", err)
		httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httpx.RespondWithJSON(w, http.StatusOK, attachments)
}
