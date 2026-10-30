package queue

import (
	"IssueForge/internal/db/sqlc"
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"go.codycody31.dev/gobullmq"
)

type NotificationCreator interface {
	CreateNotification(ctx context.Context, recepientID int64, senderID *int64, notificationType string, title, message string, projectID, issueID, workspaceID *int64) (sqlc.Notification, error)
}

type NotificationDeleter interface {
	DeleteAllReadNotifications(ctx context.Context, recepientID int64) error
	DeleteNotification(ctx context.Context, id, recepientID int64) (int64, error)
}

func NewNotificationCreateWorker(client redis.Cmdable, creator NotificationCreator) (*gobullmq.Worker[CreateNotificationJob, string], error) {
	process := func(ctx context.Context, job *gobullmq.Job[CreateNotificationJob]) (string, error) {
		data := job.Data()

		log.Printf(
			"processing notification creation: notification_id=%d recepient_id=%d issue_id=%d project_id=%d workspace_id=%d title=%s",
			data.NotificationID, data.RecepientID, data.IssueID, data.ProjectID, data.WorkspaceID, data.Title)

		if creator != nil && data.NotificationID == 0 {
			var projectID, issueID, workspaceID *int64
			if data.ProjectID > 0 {
				projectID = &data.ProjectID
			}
			if data.IssueID > 0 {
				issueID = &data.IssueID
			}
			if data.WorkspaceID > 0 {
				workspaceID = &data.WorkspaceID
			}

			notification, err := creator.CreateNotification(ctx, data.RecepientID, data.SenderID, data.NotificationType, data.Title, data.Message, projectID, issueID, workspaceID)
			if err != nil {
				return "", fmt.Errorf("create notification from queue: %w", err)
			}
			log.Printf("created notification in database: id=%d recipient=%d", notification.ID, data.RecepientID)
		}

		return "created", nil
	}

	worker, err := gobullmq.NewWorker[CreateNotificationJob, string](
		CreateNotificationQueueName,
		client,
		process,
		&gobullmq.WorkerOptions{
			Concurrency: 4,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create notification create worker: %w", err)
	}

	return worker, nil
}

func NewNotificationDeleteWorker(client redis.Cmdable, deleter NotificationDeleter) (*gobullmq.Worker[DeleteAllNotificationJob, string], error) {
	process := func(ctx context.Context, job *gobullmq.Job[DeleteAllNotificationJob]) (string, error) {
		data := job.Data()

		log.Printf(
			"processing notification deletion: notification_id=%d recepient_id=%d",
			data.NotificationID, data.RecepientID,
		)

		if deleter != nil {
			if data.NotificationID > 0 {
				if _, err := deleter.DeleteNotification(ctx, data.NotificationID, data.RecepientID); err != nil {
					return "", fmt.Errorf("delete notification %d: %w", data.NotificationID, err)
				}
			} else {
				if err := deleter.DeleteAllReadNotifications(ctx, data.RecepientID); err != nil {
					return "", fmt.Errorf("delete all read notifications for recipient %d: %w", data.RecepientID, err)
				}
			}
		}

		return "deleted", nil
	}

	worker, err := gobullmq.NewWorker[DeleteAllNotificationJob, string](
		DeleteAllNotificationQueueName, client, process, &gobullmq.WorkerOptions{
			Concurrency: 4,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create notification delete worker: %w", err)
	}

	return worker, nil
}
