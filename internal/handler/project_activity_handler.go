package handler

import (
	"IssueForge/internal/dto"
	"IssueForge/internal/httpx"
	"IssueForge/internal/middleware"
	"IssueForge/internal/service"
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type ProjectActivityService interface {
	ListProjectActivities(ctx context.Context, requesterID, projectID int64, limit, offset int32) ([]dto.ProjectActivityResponse, error)
}

type ProjectActivityHandler struct {
	projectActivityService ProjectActivityService
}

func NewProjectActivityHandler(service ProjectActivityService) *ProjectActivityHandler {
	return &ProjectActivityHandler{
		projectActivityService: service,
	}
}

func (h *ProjectActivityHandler) ListProjectActivities(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)

	projectID, err := strconv.ParseInt(vars["projectID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidProjectID.Error())
		return
	}

	limit := int32(20)
	offset := int32(0)

	if l := r.URL.Query().Get("limit"); l != "" {
		value, err := strconv.ParseInt(l, 10, 32)
		if err != nil {
			httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidLimit.Error())
			return
		}
		limit = int32(value)
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		value, err := strconv.ParseInt(o, 10, 32)
		if err != nil {
			httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidOffset.Error())
			return
		}
		offset = int32(value)
	}

	if limit <= 0 || limit > 100 {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid limit")
		return
	}
	if offset < 0 {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid offset")
		return
	}

	activities, err := h.projectActivityService.ListProjectActivities(r.Context(), requesterID, projectID, limit, offset)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidProjectID),
			errors.Is(err, service.ErrInvalidLimit),
			errors.Is(err, service.ErrInvalidOffset):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrProjectNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("list project activities fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, activities)
}
