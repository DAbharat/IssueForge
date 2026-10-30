package repository

import (
	"IssueForge/internal/db/sqlc"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type NotificationRepository struct {
	queries *sqlc.Queries
}

func NewNotificationRepository(queries *sqlc.Queries) *NotificationRepository {
	return &NotificationRepository{
		queries: queries,
	}
}

func (r *NotificationRepository) CreateNotification(ctx context.Context, recepientID int64, senderID *int64, notificationType string, title, message string, projectID, issueID, workspaceID *int64) (sqlc.Notification, error) {
	params := sqlc.CreateNotificationParams{
		RecepientID:      pgtype.Int8{Int64: recepientID, Valid: true},
		SenderID:         toInt8(senderID),
		NotificationType: sqlc.NotificationType(notificationType),
		Title:            title,
		Message:          message,
		ProjectID:        toInt8(projectID),
		IssueID:          toInt8(issueID),
		WorkspaceID:      toInt8(workspaceID),
	}

	notification, err := r.queries.CreateNotification(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				switch pgErr.ConstraintName {
				case "notifications_recepient_id_fkey", "notifications_sender_id_fkey":
					return sqlc.Notification{}, ErrUserNotFound
				case "notifications_project_id_fkey":
					return sqlc.Notification{}, ErrProjectNotFound
				case "notifications_issue_id_fkey":
					return sqlc.Notification{}, ErrIssueNotFound
				case "notifications_workspace_id_fkey":
					return sqlc.Notification{}, ErrWorkspaceNotFound
				}
			}
		}
		return sqlc.Notification{}, fmt.Errorf("create notification: %w", err)
	}

	return notification, nil
}

func (r *NotificationRepository) GetNotificationByID(ctx context.Context, id, recepientID int64) (sqlc.Notification, error) {
	params := sqlc.GetNotificationByIDParams{
		ID:          id,
		RecepientID: pgtype.Int8{Int64: recepientID, Valid: true},
	}

	notification, err := r.queries.GetNotificationByID(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Notification{}, ErrNotificationNotFound
		}
		return sqlc.Notification{}, fmt.Errorf("get notification by id: %w", err)
	}

	return notification, nil
}

func (r *NotificationRepository) ListUserNotifications(ctx context.Context, recepientID int64, limit, offset int32) ([]sqlc.Notification, error) {
	params := sqlc.ListUserNotificationsParams{
		RecepientID: pgtype.Int8{Int64: recepientID, Valid: true},
		Limit:       limit,
		Offset:      offset,
	}

	notifications, err := r.queries.ListUserNotifications(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list user notifications: %w", err)
	}

	return notifications, nil
}

func (r *NotificationRepository) ListUnreadNotifications(ctx context.Context, recepientID int64) ([]sqlc.Notification, error) {
	recepient := pgtype.Int8{Int64: recepientID, Valid: true}

	notifications, err := r.queries.ListUnreadNotifications(ctx, recepient)
	if err != nil {
		return nil, fmt.Errorf("list unread notifications: %w", err)
	}

	return notifications, nil
}

func (r *NotificationRepository) CountUnreadNotifications(ctx context.Context, recepientID int64) (int64, error) {
	recepient := pgtype.Int8{Int64: recepientID, Valid: true}

	count, err := r.queries.CountUnreadNotifications(ctx, recepient)
	if err != nil {
		return 0, fmt.Errorf("count unread notifications: %w", err)
	}

	return count, nil
}

func (r *NotificationRepository) MarkNotificationAsRead(ctx context.Context, id, recepientID int64) (sqlc.Notification, error) {
	params := sqlc.MarkNotificationAsReadParams{
		ID:          id,
		RecepientID: pgtype.Int8{Int64: recepientID, Valid: true},
	}

	notification, err := r.queries.MarkNotificationAsRead(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Notification{}, ErrNotificationNotFound
		}
		return sqlc.Notification{}, fmt.Errorf("mark notification as read: %w", err)
	}

	return notification, nil
}

func (r *NotificationRepository) MarkAllNotificationsAsRead(ctx context.Context, recepientID int64) error {
	recepient := pgtype.Int8{Int64: recepientID, Valid: true}

	if err := r.queries.MarkAllNotificationsAsRead(ctx, recepient); err != nil {
		return fmt.Errorf("mark all notifications as read: %w", err)
	}

	return nil
}

func (r *NotificationRepository) DeleteNotification(ctx context.Context, id, recepientID int64) (int64, error) {
	params := sqlc.DeleteNotificationParams{
		ID:          id,
		RecepientID: pgtype.Int8{Int64: recepientID, Valid: true},
	}

	deletedID, err := r.queries.DeleteNotification(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrNotificationNotFound
		}
		return 0, fmt.Errorf("delete notification: %w", err)
	}

	return deletedID, nil
}

func (r *NotificationRepository) DeleteAllReadNotifications(ctx context.Context, recepientID int64) error {
	recepient := pgtype.Int8{Int64: recepientID, Valid: true}

	if err := r.queries.DeleteAllReadNotifications(ctx, recepient); err != nil {
		return fmt.Errorf("delete all read notifications: %w", err)
	}

	return nil
}

func (r *NotificationRepository) DeleteOldNotifications(ctx context.Context, cutoffTime time.Time) error {
	ts := pgtype.Timestamptz{
		Time:  cutoffTime,
		Valid: true,
	}

	if err := r.queries.DeleteOldNotifications(ctx, ts); err != nil {
		return fmt.Errorf("delete old notifications: %w", err)
	}

	return nil
}

func toInt8(val *int64) pgtype.Int8 {
	if val != nil {
		return pgtype.Int8{Int64: *val, Valid: true}
	}
	return pgtype.Int8{Valid: false}
}
