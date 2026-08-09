package queue

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.codycody31.dev/gobullmq"
)

const AttachmentDeleteQueueName = "attachment-delete" //queuename

type AttachmentDeleteQueue interface {
	AddDeleteJob(ctx context.Context, job AttachmentDeleteJob) error
}

type AttachmentDeleteQueueImpl struct {
	queue *gobullmq.Queue[AttachmentDeleteJob]
}

func NewAttachmentDeleteQueue(client redis.Cmdable) (*AttachmentDeleteQueueImpl, error) {
	queue, err := gobullmq.NewQueue[AttachmentDeleteJob](
		AttachmentDeleteQueueName,
		client,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create attachment delete queue: %w", err)
	}

	return &AttachmentDeleteQueueImpl{
		queue: queue,
	}, nil
}

func (q *AttachmentDeleteQueueImpl) AddDeleteJob(ctx context.Context, job AttachmentDeleteJob) error { //jobname
	_, err := q.queue.Add(ctx, "delete-attachment", job, gobullmq.AddWithAttempts(3), gobullmq.AddWithBackoff(gobullmq.BackoffOptions{
		Type:  "exponential",
		Delay: 1000,
	}))
	if err != nil {
		return fmt.Errorf("add attachment delete job: %w", err)
	}

	return nil
}

func (q *AttachmentDeleteQueueImpl) Close() error {
	return q.queue.Close()
}
