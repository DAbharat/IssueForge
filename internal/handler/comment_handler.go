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

type CommentService interface {
	CreateComment(ctx context.Context, requesterID int64, req dto.CreateCommentRequest) (dto.CommentResponse, error)
	GetCommentByID(ctx context.Context, requesterID, commentID int64) (dto.CommentResponse, error)
	ListIssueComments(ctx context.Context, requesterID, issueID int64, limit, offset int32) ([]dto.CommentResponse, error)
	UpdateComment(ctx context.Context, requesterID, commentID int64, req dto.UpdateCommentRequest) (dto.CommentResponse, error)
	DeleteComment(ctx context.Context, requesterID, commentID int64) (int64, error)
}

type CommentHandler struct {
	commentService CommentService
}

func NewCommentHandler(service CommentService) *CommentHandler {
	return &CommentHandler{
		commentService: service,
	}
}

func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	var req dto.CreateCommentRequest

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

	issueID, err := strconv.ParseInt(vars["issueID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidIssueID.Error())
		return
	}
	req.IssueID = issueID

	comment, err := h.commentService.CreateComment(r.Context(), requesterID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidIssueID),
			errors.Is(err, service.ErrInvalidComment):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrIssueNotFound),
			errors.Is(err, service.ErrCommentNotFound),
			errors.Is(err, service.ErrUserNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("create comment fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusCreated, comment)
}

func (h *CommentHandler) GetCommentByID(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)

	commentID, err := strconv.ParseInt(vars["commentID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrCommentNotFound.Error())
		return
	}

	comment, err := h.commentService.GetCommentByID(r.Context(), requesterID, commentID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidCommentID):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrCommentNotFound),
			errors.Is(err, service.ErrIssueNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("get comment by id: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, comment)
}

func (h *CommentHandler) ListIssueComments(w http.ResponseWriter, r *http.Request) {
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

	limit := int32(20)
	offset := int32(0)

	if l := r.URL.Query().Get("limit"); l != "" {
		value, err := strconv.ParseInt(l, 10, 32)
		if err != nil {
			httpx.RespondWithError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = int32(value)
	}
	if limit <= 0 || limit > 100 {
		httpx.RespondWithError(w, http.StatusBadRequest, "limit must be between 1 and 100")
		return
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		value, err := strconv.ParseInt(o, 10, 32)
		if err != nil {
			httpx.RespondWithError(w, http.StatusBadRequest, "invalid offset")
			return
		}
		offset = int32(value)
	}
	if offset < 0 {
		httpx.RespondWithError(w, http.StatusBadRequest, "offset cannot be negative")
		return
	}

	comments, err := h.commentService.ListIssueComments(r.Context(), requesterID, issueID, limit, offset)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidCommentID),
			errors.Is(err, service.ErrInvalidComment):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrCommentNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("list issue comments fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, comments)
}

func (h *CommentHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	var req dto.UpdateCommentRequest

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

	commentID, err := strconv.ParseInt(vars["commentID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidCommentID.Error())
		return
	}

	comment, err := h.commentService.UpdateComment(r.Context(), requesterID, commentID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidCommentID),
			errors.Is(err, service.ErrInvalidComment):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrCommentNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("update comment fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, comment)
}

func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
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

	id, err := h.commentService.DeleteComment(r.Context(), requesterID, commentID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidCommentID):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrCommentNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("delete comment fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, id)
}
