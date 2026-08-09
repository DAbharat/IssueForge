package queue

import (
	"IssueForge/internal/storage"
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"go.codycody31.dev/gobullmq"
)

func NewAttachmentDeleteWorker(client redis.Cmdable, storage storage.Storage) (*gobullmq.Worker[AttachmentDeleteJob, string], error) {
	process := func(ctx context.Context, job *gobullmq.Job[AttachmentDeleteJob]) (string, error) {
		data := job.Data()

		log.Printf(
			"processing attachment deletion: attachment%d issue=%d publicID=%s",
			data.AttachmentID,
			data.IssueID,
			data.FilePublicID,
		)
		if err := storage.Delete(ctx, data.FilePublicID, data.ResourceType); err != nil {
			return "", fmt.Errorf("delete attachment from storage: %w", err)
		}
		return "deleted", nil
	}

	worker, err := gobullmq.NewWorker[AttachmentDeleteJob, string](
		AttachmentDeleteQueueName,
		client,
		process,
		&gobullmq.WorkerOptions{
			Concurrency: 4,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create attachment delete worker: %w", err)
	}

	return worker, nil
}
