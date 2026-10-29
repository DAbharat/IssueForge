package handler

import (
	"IssueForge/internal/dto"
	"IssueForge/internal/httpx"
	"IssueForge/internal/middleware"
	"IssueForge/internal/service"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type AnalyticsService interface {
	GetWorkspaceAnalytics(
		ctx context.Context,
		requesterID, workspaceID int64,
		from, to *time.Time,
		memberLimit int32,
	) (dto.WorkspaceAnalyticsResponse, error)
	GetProjectAnalytics(
		ctx context.Context,
		requesterID, projectID int64,
		from, to *time.Time,
		memberLimit int32,
	) (dto.ProjectAnalyticsResponse, error)
}

type AnalyticsHandler struct {
	analyticsService AnalyticsService
}

func NewAnalyticsHandler(service AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsService: service}
}

func parseDateQuery(val string) (*time.Time, error) {
	if val == "" {
		return nil, nil
	}
	if t, err := time.Parse(time.RFC3339, val); err == nil {
		return &t, nil
	}
	if t, err := time.Parse("2006-01-02", val); err == nil {
		return &t, nil
	}
	return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD or RFC3339")
}

func (h *AnalyticsHandler) GetWorkspaceAnalytics(w http.ResponseWriter, r *http.Request) {
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

	from, err := parseDateQuery(r.URL.Query().Get("from"))
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	to, err := parseDateQuery(r.URL.Query().Get("to"))
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	limit := int32(10)
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.ParseInt(l, 10, 32); err == nil && val > 0 && val <= 50 {
			limit = int32(val)
		}
	}

	res, err := h.analyticsService.GetWorkspaceAnalytics(r.Context(), requesterID, workspaceID, from, to, limit)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrWorkspaceNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrInvalidWorkspaceID):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		default:
			log.Printf("get workspace analytics fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	httpx.RespondWithJSON(w, http.StatusOK, res)
}

func (h *AnalyticsHandler) GetProjectAnalytics(w http.ResponseWriter, r *http.Request) {
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

	from, err := parseDateQuery(r.URL.Query().Get("from"))
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	to, err := parseDateQuery(r.URL.Query().Get("to"))
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	limit := int32(10)
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.ParseInt(l, 10, 32); err == nil && val > 0 && val <= 50 {
			limit = int32(val)
		}
	}

	res, err := h.analyticsService.GetProjectAnalytics(r.Context(), requesterID, projectID, from, to, limit)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrProjectNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrInvalidProjectID):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		default:
			log.Printf("get project analytics fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	httpx.RespondWithJSON(w, http.StatusOK, res)
}
