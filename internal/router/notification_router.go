package router

import (
	"IssueForge/internal/handler"
	"IssueForge/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func registerNotificationRoutes(
	r *mux.Router,
	notificationHandler *handler.NotificationHandler,
	authMiddleware *middleware.AuthMiddleware,
	readRateLimit *middleware.RateLimitMiddleware,
) {
	r.Handle("/api/notifications",
		authMiddleware.Authenticate(
			readRateLimit.Limit(http.HandlerFunc(notificationHandler.ListNotifications)),
		),
	).Methods("GET")

	r.Handle("/api/notifications/unread",
		authMiddleware.Authenticate(
			readRateLimit.Limit(http.HandlerFunc(notificationHandler.ListUnread)),
		),
	).Methods("GET")

	r.Handle("/api/notifications/unread/count",
		authMiddleware.Authenticate(
			readRateLimit.Limit(http.HandlerFunc(notificationHandler.CountUnread)),
		),
	).Methods("GET")

	r.Handle("/api/notifications/{id:[0-9]+}",
		authMiddleware.Authenticate(
			readRateLimit.Limit(http.HandlerFunc(notificationHandler.GetNotification)),
		),
	).Methods("GET")

	r.Handle("/api/notifications/{id:[0-9]+}/read",
		authMiddleware.Authenticate(
			http.HandlerFunc(notificationHandler.MarkAsRead),
		),
	).Methods("PATCH")

	r.Handle("/api/notifications/read-all",
		authMiddleware.Authenticate(
			http.HandlerFunc(notificationHandler.MarkAllAsRead),
		),
	).Methods("POST")

	r.Handle("/api/notifications/{id:[0-9]+}",
		authMiddleware.Authenticate(
			http.HandlerFunc(notificationHandler.DeleteNotification),
		),
	).Methods("DELETE")

	r.Handle("/api/notifications",
		authMiddleware.Authenticate(
			http.HandlerFunc(notificationHandler.DeleteAllNotifications),
		),
	).Methods("DELETE")
}
