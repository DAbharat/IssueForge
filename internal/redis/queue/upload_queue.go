package queue

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.codycody31.dev/gobullmq"
)

const AttachmentUploadQueueName = "attachment-upload"

func NewAttachmentUploadQueue(client redis.Cmdable) (*gobullmq.Queue[AttachmentJob], error) {
	q, err := gobullmq.NewQueue[AttachmentJob](
		AttachmentUploadQueueName,
		client,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create attachment upload queue: %w", err)
	}
	return q, nil
}

func AddAttachmentUploadJob(ctx context.Context, q *gobullmq.Queue[AttachmentJob], job AttachmentJob) error {
	_, err := q.Add(
		ctx,
		"upload-attachment",
		job,
	)
	if err != nil {
		return fmt.Errorf("add attachment upload job: %w", err)
	}
	return nil
}
