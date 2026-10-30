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

type NotificationService interface {
	GetNotificationByID(ctx context.Context, requesterID, id int64) (dto.NotificationResponse, error)
	ListUserNotifications(ctx context.Context, requesterID int64, limit, offset int32) ([]dto.NotificationResponse, error)
	ListUnreadNotifications(ctx context.Context, requesterID int64) ([]dto.NotificationResponse, error)
	CountUnreadNotifications(ctx context.Context, requesterID int64) (dto.UnreadCountResponse, error)
	MarkNotificationAsRead(ctx context.Context, requesterID, id int64) (dto.NotificationResponse, error)
	MarkAllNotificationsAsRead(ctx context.Context, requesterID int64) error
	DeleteNotification(ctx context.Context, requesterID, id int64) error
	DeleteAllNotifications(ctx context.Context, requesterID int64) error
}

type NotificationHandler struct {
	notificationService NotificationService
}

func NewNotificationHandler(service NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: service,
	}
}

func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	filter := r.URL.Query().Get("filter")
	if filter == "unread" {
		notifications, err := h.notificationService.ListUnreadNotifications(r.Context(), requesterID)
		if err != nil {
			log.Printf("list unread notifications fail: %v", err)
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		httpx.RespondWithJSON(w, http.StatusOK, notifications)
		return
	}

	limit := int32(20)
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = int32(val)
		}
	}

	offset := int32(0)
	if o := r.URL.Query().Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = int32(val)
		}
	}

	notifications, err := h.notificationService.ListUserNotifications(r.Context(), requesterID, limit, offset)
	if err != nil {
		log.Printf("list user notifications fail: %v", err)
		httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httpx.RespondWithJSON(w, http.StatusOK, notifications)
}

func (h *NotificationHandler) ListUnread(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	notifications, err := h.notificationService.ListUnreadNotifications(r.Context(), requesterID)
	if err != nil {
		log.Printf("list unread notifications fail: %v", err)
		httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httpx.RespondWithJSON(w, http.StatusOK, notifications)
}

func (h *NotificationHandler) CountUnread(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	res, err := h.notificationService.CountUnreadNotifications(r.Context(), requesterID)
	if err != nil {
		log.Printf("count unread notifications fail: %v", err)
		httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httpx.RespondWithJSON(w, http.StatusOK, res)
}

func (h *NotificationHandler) GetNotification(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid notification id")
		return
	}

	notification, err := h.notificationService.GetNotificationByID(r.Context(), requesterID, id)
	if err != nil {
		if errors.Is(err, service.ErrNotificationNotFound) {
			httpx.RespondWithError(w, http.StatusNotFound, "notification not found")
			return
		}
		log.Printf("get notification fail: %v", err)
		httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httpx.RespondWithJSON(w, http.StatusOK, notification)
}

func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid notification id")
		return
	}

	notification, err := h.notificationService.MarkNotificationAsRead(r.Context(), requesterID, id)
	if err != nil {
		if errors.Is(err, service.ErrNotificationNotFound) {
			httpx.RespondWithError(w, http.StatusNotFound, "notification not found")
			return
		}
		log.Printf("mark notification read fail: %v", err)
		httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httpx.RespondWithJSON(w, http.StatusOK, notification)
}

func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.notificationService.MarkAllNotificationsAsRead(r.Context(), requesterID); err != nil {
		log.Printf("mark all notifications read fail: %v", err)
		httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httpx.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "all notifications marked as read"})
}

func (h *NotificationHandler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		httpx.RespondWithError(w, http.StatusBadRequest, "invalid notification id")
		return
	}

	if err := h.notificationService.DeleteNotification(r.Context(), requesterID, id); err != nil {
		if errors.Is(err, service.ErrNotificationNotFound) {
			httpx.RespondWithError(w, http.StatusNotFound, "notification not found")
			return
		}
		log.Printf("delete notification fail: %v", err)
		httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httpx.RespondWithJSON(w, http.StatusOK, map[string]int64{"id": id})
}

func (h *NotificationHandler) DeleteAllNotifications(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		httpx.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.notificationService.DeleteAllNotifications(r.Context(), requesterID); err != nil {
		log.Printf("delete all notifications fail: %v", err)
		httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httpx.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "all notifications queued for deletion"})
}
