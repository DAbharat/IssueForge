package cache

import (
	"IssueForge/internal/dto"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type WorkspaceCache interface {
	GetWorkspace(ctx context.Context, workspaceID int64) (dto.WorkspaceResponse, bool, error)
	SetWorkspace(ctx context.Context, workspace dto.WorkspaceResponse) error
	DeleteWorkspace(ctx context.Context, workspaceID int64) error
}

type RedisWorkspaceCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisWorkspaceCache(client *redis.Client, ttl time.Duration) *RedisWorkspaceCache {
	if ttl == 0 {
		ttl = 10 * time.Minute
	}
	return &RedisWorkspaceCache{
		client: client,
		ttl:    ttl,
	}
}

func workspaceKey(workspaceID int64) string {
	return fmt.Sprintf("workspace:%d", workspaceID)
}

func (c *RedisWorkspaceCache) GetWorkspace(ctx context.Context, workspaceID int64) (dto.WorkspaceResponse, bool, error) {
	var workspace dto.WorkspaceResponse

	data, err := c.client.Get(ctx, workspaceKey(workspaceID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return dto.WorkspaceResponse{}, false, nil
		}
		return dto.WorkspaceResponse{}, false, err
	}

	if err := json.Unmarshal([]byte(data), &workspace); err != nil {
		_ = c.DeleteWorkspace(ctx, workspaceID)
		return dto.WorkspaceResponse{}, false, err
	}
	return workspace, true, nil
}

func (c *RedisWorkspaceCache) SetWorkspace(ctx context.Context, workspace dto.WorkspaceResponse) error {
	if workspace.ID <= 0 {
		return errors.New("invalid workspace id")
	}

	data, err := json.Marshal(workspace)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, workspaceKey(workspace.ID), data, c.ttl).Err()
}

func (c *RedisWorkspaceCache) DeleteWorkspace(ctx context.Context, workspaceID int64) error {
	if workspaceID <= 0 {
		return errors.New("invalid workspace id")
	}

	return c.client.Del(ctx, workspaceKey(workspaceID)).Err()
}
