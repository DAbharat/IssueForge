package handler

import (
	"IssueForge/internal/dto"
	"IssueForge/internal/httpx"
	"IssueForge/internal/middleware"
	"IssueForge/internal/service"
	"context"
	"errors"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type IssueAttachmentsService interface {
	CreateAttachment(ctx context.Context, requesterID, issueID int64, commentID *int64, file multipart.File, header *multipart.FileHeader) (dto.AttachmentResponse, error)
	GetAttachmentByID(ctx context.Context, requesterID, id int64) (dto.AttachmentResponse, error)
	ListCommentAttachments(ctx context.Context, requesterID, commentID int64) ([]dto.AttachmentResponse, error)
	ListIssueAttachments(ctx context.Context, requesterID, issueID int64) ([]dto.AttachmentResponse, error)
	SoftDeleteAttachments(ctx context.Context, requesterID, id int64) (int64, error)
}

type IssueAttachmentsHandler struct {
	issueAttachmentsService IssueAttachmentsService
}

func NewIssueAttachmentsHandler(service IssueAttachmentsService) *IssueAttachmentsHandler {
	return &IssueAttachmentsHandler{
		issueAttachmentsService: service,
	}
}

func (h *IssueAttachmentsHandler) CreateAttachment(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 11<<20)

	if err := r.ParseMultipartForm(11 << 20); err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()

	vars := mux.Vars(r)

	issueID, err := strconv.ParseInt(vars["issueID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidIssueID.Error())
		return
	}

	var commentID *int64
	if value := r.FormValue("comment_id"); value != "" {
		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidCommentID.Error())
			return
		}
		commentID = &id
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	attachment, err := h.issueAttachmentsService.CreateAttachment(r.Context(), requesterID, issueID, commentID, file, header)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidIssueID),
			errors.Is(err, service.ErrInvalidCommentID):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrEmptyFile),
			errors.Is(err, service.ErrFileTooLarge),
			errors.Is(err, service.ErrUnsupportedType):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrIssueNotFound),
			errors.Is(err, service.ErrCommentNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("create attachment fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusCreated, attachment)
}

func (h *IssueAttachmentsHandler) GetAttachmentByID(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)

	attachmentID, err := strconv.ParseInt(vars["attachmentID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidAttachmentID.Error())
		return
	}

	attachment, err := h.issueAttachmentsService.GetAttachmentByID(r.Context(), requesterID, attachmentID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrAttachmentNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("get attachment by id fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, attachment)
}

func (h *IssueAttachmentsHandler) ListIssueAttachments(w http.ResponseWriter, r *http.Request) {
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

	attachment, err := h.issueAttachmentsService.ListIssueAttachments(r.Context(), requesterID, issueID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidIssueID):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrIssueNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("list issue attachments fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, attachment)
}

func (h *IssueAttachmentsHandler) ListCommentAttachments(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)

	commentID, err := strconv.ParseInt(vars["commentID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidCommentID.Error())
		return
	}

	attachment, err := h.issueAttachmentsService.ListCommentAttachments(r.Context(), requesterID, commentID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidCommentID):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrCommentNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("list comment attachments: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, attachment)
}

func (h *IssueAttachmentsHandler) SoftDeleteAttachments(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)

	attachmentID, err := strconv.ParseInt(vars["attachmentID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidAttachmentID.Error())
		return
	}

	attachment, err := h.issueAttachmentsService.SoftDeleteAttachments(r.Context(), requesterID, attachmentID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrAttachmentNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrDeleteFailed):
			httpx.RespondWithError(w, http.StatusInternalServerError, err.Error())
		default:
			log.Printf("soft delete attachment failed: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, map[string]int64{
		"attachmentID": attachment,
	})
}
