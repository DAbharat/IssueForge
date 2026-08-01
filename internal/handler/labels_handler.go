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

type LabelsService interface {
	CreateLabel(ctx context.Context, requesterID, projectID int64, req dto.CreateLabelRequest) (dto.LabelResponse, error)
	GetLabelByID(ctx context.Context, requesterID, labelID int64) (dto.LabelResponse, error)
	ListProjectLabels(ctx context.Context, requesterID, projectID int64) ([]dto.LabelResponse, error)
	UpdateLabel(ctx context.Context, requesterID int64, req dto.UpdateLabelRequest, labelID int64) (dto.LabelResponse, error)
	DeleteLabel(ctx context.Context, requesterID, labelID int64) (dto.LabelResponse, error)
	AttachLabelsToIssue(ctx context.Context, requesterID, issueID int64, req dto.AttachLabelsRequest) error
	RemoveLabelFromIssue(ctx context.Context, requesterID, issueID, labelID int64) (int64, error)
	ListIssueLabels(ctx context.Context, requesterID, issueID int64) ([]dto.LabelResponse, error)
}

type LabelsHandler struct {
	labelsService LabelsService
}

func NewLabelsHandler(service LabelsService) *LabelsHandler {
	return &LabelsHandler{
		labelsService: service,
	}
}

func (h *LabelsHandler) CreateLabel(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.CreateLabelRequest

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid or oversized request body")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		httpx.RespondWithError(w, http.StatusBadRequest, "request body must contain only a single json object")
		return
	}

	vars := mux.Vars(r)

	projectID, err := strconv.ParseInt(vars["projectID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidProjectID.Error())
		return
	}

	label, err := h.labelsService.CreateLabel(r.Context(), requesterID, projectID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidProjectID),
			errors.Is(err, service.ErrInvalidLabelName),
			errors.Is(err, service.ErrInvalidColor):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrUserNotFound),
			errors.Is(err, service.ErrProjectNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrLabelAlreadyExists):
			httpx.RespondWithError(w, http.StatusConflict, err.Error())
		default:
			log.Printf("create label fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusCreated, label)
}

func (h *LabelsHandler) GetLabelByID(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)

	labelID, err := strconv.ParseInt(vars["labelID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidLabelID.Error())
		return
	}

	label, err := h.labelsService.GetLabelByID(r.Context(), requesterID, labelID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidLabelID):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrLabelNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("get label by id fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, label)
}

func (h *LabelsHandler) ListProjectLabels(w http.ResponseWriter, r *http.Request) {
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

	label, err := h.labelsService.ListProjectLabels(r.Context(), requesterID, projectID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidProjectID):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		default:
			log.Printf("list project labels: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, label)
}

func (h *LabelsHandler) UpdateLabel(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.UpdateLabelRequest

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

	labelID, err := strconv.ParseInt(vars["labelID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidLabelID.Error())
		return
	}

	label, err := h.labelsService.UpdateLabel(r.Context(), requesterID, req, labelID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidLabelID),
			errors.Is(err, service.ErrInvalidLabelName),
			errors.Is(err, service.ErrInvalidColor):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrLabelNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrLabelAlreadyExists):
			httpx.RespondWithError(w, http.StatusConflict, err.Error())
		default:
			log.Printf("update label fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, label)
}

func (h *LabelsHandler) DeleteLabel(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)

	labelID, err := strconv.ParseInt(vars["labelID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidLabelID.Error())
		return
	}

	label, err := h.labelsService.DeleteLabel(r.Context(), requesterID, labelID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidLabelID):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrLabelNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("delete label fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, label)
}

func (h *LabelsHandler) AttachLabelsToIssue(w http.ResponseWriter, r *http.Request) {
	requestID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req dto.AttachLabelsRequest

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid or oversized request body")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		httpx.RespondWithError(w, http.StatusBadRequest, "request body must only contain a single json object")
		return
	}

	vars := mux.Vars(r)

	issueID, err := strconv.ParseInt(vars["issueID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidIssueID.Error())
		return
	}

	err = h.labelsService.AttachLabelsToIssue(r.Context(), requestID, issueID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidIssueID),
			errors.Is(err, service.ErrInvalidLabelID),
			errors.Is(err, service.ErrLabelDoesNotBelongToProject):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrIssueNotFound),
			errors.Is(err, service.ErrLabelNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("attach labels to issue: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "labels attached successfully",
	})
}

func (h *LabelsHandler) RemoveLabelFromIssue(w http.ResponseWriter, r *http.Request) {
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

	labelID, err := strconv.ParseInt(vars["labelID"], 10, 64)
	if err != nil {
		httpx.RespondWithError(w, http.StatusBadRequest, service.ErrInvalidLabelID.Error())
		return
	}

	label, err := h.labelsService.RemoveLabelFromIssue(r.Context(), requesterID, issueID, labelID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidIssueID),
			errors.Is(err, service.ErrInvalidLabelID):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrIssueNotFound),
			errors.Is(err, service.ErrLabelNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("remove label from issue fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, map[string]int64{
		"label_id": label,
	})
}

func (h *LabelsHandler) ListIssueLabels(w http.ResponseWriter, r *http.Request) {
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

	label, err := h.labelsService.ListIssueLabels(r.Context(), requesterID, issueID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			httpx.RespondWithError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrInvalidIssueID):
			httpx.RespondWithError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrIssueNotFound):
			httpx.RespondWithError(w, http.StatusNotFound, err.Error())
		default:
			log.Printf("list issue labels fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	httpx.RespondWithJSON(w, http.StatusOK, label)
}
