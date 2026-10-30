package queue

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.codycody31.dev/gobullmq"
)

const CreateNotificationQueueName = "create-notification"
const DeleteAllNotificationQueueName = "delete-all-notification"

type CreateNotificationQueue interface {
	AddCreateJob(ctx context.Context, job CreateNotificationJob) error
}

type DeleteAllNotificationQueue interface {
	AddDeleteAllJob(ctx context.Context, job DeleteAllNotificationJob) error
}

type CreateNotificationQueueImpl struct {
	queue *gobullmq.Queue[CreateNotificationJob]
}

type DeleteAllNotificationQueueImpl struct {
	queue *gobullmq.Queue[DeleteAllNotificationJob]
}

func NewCreateNotificationQueue(client redis.Cmdable) (*CreateNotificationQueueImpl, error) {
	queue, err := gobullmq.NewQueue[CreateNotificationJob](
		CreateNotificationQueueName, client, nil,
	)

	if err != nil {
		return nil, fmt.Errorf("create notification queue: %w", err)
	}

	return &CreateNotificationQueueImpl{
		queue: queue,
	}, nil
}

func NewDeleteAllNotificationsQueue(client redis.Cmdable) (*DeleteAllNotificationQueueImpl, error) {
	queue, err := gobullmq.NewQueue[DeleteAllNotificationJob](
		DeleteAllNotificationQueueName, client, nil,
	)

	if err != nil {
		return nil, fmt.Errorf("delete all notification queue: %w", err)
	}

	return &DeleteAllNotificationQueueImpl{
		queue: queue,
	}, nil
}

func (q *CreateNotificationQueueImpl) AddCreateJob(ctx context.Context, job CreateNotificationJob) error {
	_, err := q.queue.Add(ctx, "notification-create", job, gobullmq.AddWithAttempts(3), gobullmq.AddWithBackoff(gobullmq.BackoffOptions{
		Type:  "exponential",
		Delay: 1000,
	}))
	if err != nil {
		return fmt.Errorf("add create notification job: %w", err)
	}

	return nil
}

func (q *DeleteAllNotificationQueueImpl) AddDeleteAllJob(ctx context.Context, job DeleteAllNotificationJob) error {
	_, err := q.queue.Add(ctx, "all-notification-delete", job, gobullmq.AddWithAttempts(3), gobullmq.AddWithBackoff(gobullmq.BackoffOptions{
		Type:  "exponential",
		Delay: 1000,
	}))
	if err != nil {
		return fmt.Errorf("add all delete notification job: %w", err)
	}

	return nil
}

func (q *CreateNotificationQueueImpl) Close() error {
	return q.queue.Close()
}

func (q *DeleteAllNotificationQueueImpl) Close() error {
	return q.queue.Close()
}
