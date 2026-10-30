package service

import (
	"IssueForge/internal/db/sqlc"
	"IssueForge/internal/dto"
	"IssueForge/internal/redis/queue"
	"IssueForge/internal/repository"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type NotificationRepo interface {
	CreateNotification(ctx context.Context, recepientID int64, senderID *int64, notificationType string, title, message string, projectID, issueID, workspaceID *int64) (sqlc.Notification, error)
	GetNotificationByID(ctx context.Context, id, recepientID int64) (sqlc.Notification, error)
	ListUserNotifications(ctx context.Context, recepientID int64, limit, offset int32) ([]sqlc.Notification, error)
	ListUnreadNotifications(ctx context.Context, recepientID int64) ([]sqlc.Notification, error)
	CountUnreadNotifications(ctx context.Context, recepientID int64) (int64, error)
	MarkNotificationAsRead(ctx context.Context, id, recepientID int64) (sqlc.Notification, error)
	MarkAllNotificationsAsRead(ctx context.Context, recepientID int64) error
	DeleteNotification(ctx context.Context, id, recepientID int64) (int64, error)
	DeleteAllReadNotifications(ctx context.Context, recepientID int64) error
	DeleteOldNotifications(ctx context.Context, cutoffTime time.Time) error
}

type NotificationService struct {
	repo           NotificationRepo
	createQueue    queue.CreateNotificationQueue
	deleteAllQueue queue.DeleteAllNotificationQueue
}

func NewNotificationService(repo NotificationRepo, createQueue queue.CreateNotificationQueue, deleteAllQueue queue.DeleteAllNotificationQueue) *NotificationService {
	return &NotificationService{
		repo:           repo,
		createQueue:    createQueue,
		deleteAllQueue: deleteAllQueue,
	}
}

func (s *NotificationService) CreateNotification(ctx context.Context, req dto.CreateNotificationRequest) (dto.NotificationResponse, error) {
	if req.RecepientID <= 0 {
		return dto.NotificationResponse{}, ErrInvalidUserID
	}
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		return dto.NotificationResponse{}, ErrInvalidTitle
	}
	req.Message = strings.TrimSpace(req.Message)
	if req.Message == "" {
		return dto.NotificationResponse{}, errors.New("notification message cannot be empty")
	}
	req.NotificationType = strings.TrimSpace(req.NotificationType)
	if req.NotificationType == "" {
		return dto.NotificationResponse{}, errors.New("notification type is required")
	}

	notification, err := s.repo.CreateNotification(ctx, req.RecepientID, req.SenderID, req.NotificationType, req.Title, req.Message, req.ProjectID, req.IssueID, req.WorkspaceID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return dto.NotificationResponse{}, ErrUserNotFound
		}
		if errors.Is(err, repository.ErrProjectNotFound) {
			return dto.NotificationResponse{}, ErrProjectNotFound
		}
		if errors.Is(err, repository.ErrIssueNotFound) {
			return dto.NotificationResponse{}, ErrIssueNotFound
		}
		if errors.Is(err, repository.ErrWorkspaceNotFound) {
			return dto.NotificationResponse{}, ErrWorkspaceNotFound
		}
		return dto.NotificationResponse{}, fmt.Errorf("create notification: %w", err)
	}

	return toNotificationResponse(notification), nil
}

func (s *NotificationService) EnqueueCreateNotification(ctx context.Context, job queue.CreateNotificationJob) error {
	if job.RecepientID <= 0 {
		return ErrInvalidUserID
	}
	if s.createQueue == nil {
		return errors.New("create notification queue is not configured")
	}

	if err := s.createQueue.AddCreateJob(ctx, job); err != nil {
		return fmt.Errorf("enqueue notification creation: %w", err)
	}

	return nil
}

type NotificationSender interface {
	SendNotification(
		ctx context.Context,
		recepientID int64,
		senderID *int64,
		notificationType string,
		title, message string,
		projectID, issueID, workspaceID *int64,
	) error
}

func (s *NotificationService) SendNotification(
	ctx context.Context,
	recepientID int64,
	senderID *int64,
	notificationType string,
	title, message string,
	projectID, issueID, workspaceID *int64,
) error {
	if recepientID <= 0 {
		return nil
	}

	// Do not notify self
	if senderID != nil && *senderID == recepientID {
		return nil
	}

	if s.createQueue != nil {
		var pID, iID, wID int64
		if projectID != nil {
			pID = *projectID
		}
		if issueID != nil {
			iID = *issueID
		}
		if workspaceID != nil {
			wID = *workspaceID
		}

		job := queue.CreateNotificationJob{
			RecepientID:      recepientID,
			SenderID:         senderID,
			NotificationType: notificationType,
			Title:            title,
			Message:          message,
			ProjectID:        pID,
			IssueID:          iID,
			WorkspaceID:      wID,
		}
		if err := s.createQueue.AddCreateJob(ctx, job); err != nil {
			return fmt.Errorf("enqueue notification: %w", err)
		}
		return nil
	}

	_, err := s.repo.CreateNotification(ctx, recepientID, senderID, notificationType, title, message, projectID, issueID, workspaceID)
	return err
}

func (s *NotificationService) GetNotificationByID(ctx context.Context, requesterID, id int64) (dto.NotificationResponse, error) {
	if requesterID <= 0 {
		return dto.NotificationResponse{}, ErrInvalidUserID
	}
	if id <= 0 {
		return dto.NotificationResponse{}, ErrInvalidNotificationID
	}

	notification, err := s.repo.GetNotificationByID(ctx, id, requesterID)
	if err != nil {
		if errors.Is(err, repository.ErrNotificationNotFound) {
			return dto.NotificationResponse{}, ErrNotificationNotFound
		}
		return dto.NotificationResponse{}, fmt.Errorf("get notification: %w", err)
	}

	return toNotificationResponse(notification), nil
}

func (s *NotificationService) ListUserNotifications(ctx context.Context, requesterID int64, limit, offset int32) ([]dto.NotificationResponse, error) {
	if requesterID <= 0 {
		return nil, ErrInvalidUserID
	}

	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	notifications, err := s.repo.ListUserNotifications(ctx, requesterID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list user notifications: %w", err)
	}

	res := make([]dto.NotificationResponse, 0, len(notifications))
	for _, n := range notifications {
		res = append(res, toNotificationResponse(n))
	}

	return res, nil
}

func (s *NotificationService) ListUnreadNotifications(ctx context.Context, requesterID int64) ([]dto.NotificationResponse, error) {
	if requesterID <= 0 {
		return nil, ErrInvalidUserID
	}

	notifications, err := s.repo.ListUnreadNotifications(ctx, requesterID)
	if err != nil {
		return nil, fmt.Errorf("list unread notifications: %w", err)
	}

	res := make([]dto.NotificationResponse, 0, len(notifications))
	for _, n := range notifications {
		res = append(res, toNotificationResponse(n))
	}

	return res, nil
}

func (s *NotificationService) CountUnreadNotifications(ctx context.Context, requesterID int64) (dto.UnreadCountResponse, error) {
	if requesterID <= 0 {
		return dto.UnreadCountResponse{}, ErrInvalidUserID
	}

	count, err := s.repo.CountUnreadNotifications(ctx, requesterID)
	if err != nil {
		return dto.UnreadCountResponse{}, fmt.Errorf("count unread notifications: %w", err)
	}

	return dto.UnreadCountResponse{UnreadCount: count}, nil
}

func (s *NotificationService) MarkNotificationAsRead(ctx context.Context, requesterID, id int64) (dto.NotificationResponse, error) {
	if requesterID <= 0 {
		return dto.NotificationResponse{}, ErrInvalidUserID
	}
	if id <= 0 {
		return dto.NotificationResponse{}, ErrInvalidNotificationID
	}

	notification, err := s.repo.MarkNotificationAsRead(ctx, id, requesterID)
	if err != nil {
		if errors.Is(err, repository.ErrNotificationNotFound) {
			return dto.NotificationResponse{}, ErrNotificationNotFound
		}
		return dto.NotificationResponse{}, fmt.Errorf("mark notification as read: %w", err)
	}

	return toNotificationResponse(notification), nil
}

func (s *NotificationService) MarkAllNotificationsAsRead(ctx context.Context, requesterID int64) error {
	if requesterID <= 0 {
		return ErrInvalidUserID
	}

	if err := s.repo.MarkAllNotificationsAsRead(ctx, requesterID); err != nil {
		return fmt.Errorf("mark all notifications as read: %w", err)
	}

	return nil
}

func (s *NotificationService) DeleteNotification(ctx context.Context, requesterID, id int64) error {
	if requesterID <= 0 {
		return ErrInvalidUserID
	}
	if id <= 0 {
		return ErrInvalidNotificationID
	}

	_, err := s.repo.DeleteNotification(ctx, id, requesterID)
	if err != nil {
		if errors.Is(err, repository.ErrNotificationNotFound) {
			return ErrNotificationNotFound
		}
		return fmt.Errorf("delete notification: %w", err)
	}

	return nil
}

func (s *NotificationService) DeleteAllNotifications(ctx context.Context, requesterID int64) error {
	if requesterID <= 0 {
		return ErrInvalidUserID
	}

	if err := s.repo.MarkAllNotificationsAsRead(ctx, requesterID); err != nil {
		return fmt.Errorf("mark all as read before deletion: %w", err)
	}

	if s.deleteAllQueue != nil {
		job := queue.DeleteAllNotificationJob{
			RecepientID: requesterID,
		}
		if err := s.deleteAllQueue.AddDeleteAllJob(ctx, job); err != nil {
			return fmt.Errorf("enqueue delete all notifications: %w", err)
		}
	} else {
		if err := s.repo.DeleteAllReadNotifications(ctx, requesterID); err != nil {
			return fmt.Errorf("delete all read notifications: %w", err)
		}
	}

	return nil
}

func (s *NotificationService) DeleteOldNotifications(ctx context.Context, cutoffTime time.Time) error {
	if err := s.repo.DeleteOldNotifications(ctx, cutoffTime); err != nil {
		return fmt.Errorf("delete old notifications: %w", err)
	}
	return nil
}

func toNotificationResponse(n sqlc.Notification) dto.NotificationResponse {
	var senderID, projectID, issueID, workspaceID *int64
	var readAt *time.Time

	if n.SenderID.Valid {
		id := n.SenderID.Int64
		senderID = &id
	}
	if n.ProjectID.Valid {
		id := n.ProjectID.Int64
		projectID = &id
	}
	if n.IssueID.Valid {
		id := n.IssueID.Int64
		issueID = &id
	}
	if n.WorkspaceID.Valid {
		id := n.WorkspaceID.Int64
		workspaceID = &id
	}
	if n.ReadAt.Valid {
		t := n.ReadAt.Time
		readAt = &t
	}

	return dto.NotificationResponse{
		ID:               n.ID,
		RecepientID:      n.RecepientID.Int64,
		SenderID:         senderID,
		NotificationType: string(n.NotificationType),
		Title:            n.Title,
		Message:          n.Message,
		ProjectID:        projectID,
		IssueID:          issueID,
		WorkspaceID:      workspaceID,
		IsRead:           n.IsRead,
		ReadAt:           readAt,
		CreatedAt:        n.CreatedAt.Time,
		UpdatedAt:        n.UpdatedAt.Time,
	}
}
