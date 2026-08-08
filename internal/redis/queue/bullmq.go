package queue

import (
	"context"
	"fmt"

	"github.com/ktbsomen/gobullmq"
	"github.com/redis/go-redis/v9"
)

func NewAttachmentQueue(ctx context.Context, client *redis.Client) (*gobullmq.Queue, error) {
	queue, err := gobullmq.NewQueue(
		ctx, "attachment", client,
	)
	if err != nil {
		return nil, fmt.Errorf("create attachment queue: %w", err)
	}

	return queue, nil
}
