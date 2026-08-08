package queue

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"go.codycody31.dev/gobullmq"
)

type AttachmentUploadProcessor interface {
	ProcessUpload(ctx context.Context, job AttachmentJob) error
}

func NewAttachmentUploadWorker(client redis.Cmdable, processor AttachmentUploadProcessor) (*gobullmq.Worker[AttachmentJob, string], error) {
	process := func(ctx context.Context, job *gobullmq.Job[AttachmentJob]) (string, error) {
		data := job.Data()

		log.Printf(
			"processing attachment upload: issue=%d user=%d publicID=%s",
			data.IssueID,
			data.UserID,
			data.FilePublicID,
		)
		if err := processor.ProcessUpload(ctx, data); err != nil {
			return "", fmt.Errorf("process attachment upload: %w", err)
		}
		return "uploaded", nil
	}

	worker, err := gobullmq.NewWorker[AttachmentJob, string](
		AttachmentUploadQueueName,
		client,
		process,
		&gobullmq.WorkerOptions{
			Concurrency: 4,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create attachment upload worker: %w", err)
	}
	return worker, nil
}
