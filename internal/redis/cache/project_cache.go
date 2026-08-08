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

type ProjectCache interface {
	GetProject(ctx context.Context, projectID int64) (dto.ProjectResponse, bool, error)
	SetProject(ctx context.Context, project dto.ProjectResponse) error
	DeleteProject(ctx context.Context, projectID int64) error
}

type RedisProjectCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisProjectCache(client *redis.Client, ttl time.Duration) *RedisProjectCache {
	if ttl == 0 {
		ttl = 10 * time.Minute
	}
	return &RedisProjectCache{
		client: client,
		ttl:    ttl,
	}
}

func projectKey(projectID int64) string {
	return fmt.Sprintf("project:%d", projectID)
}

func (c *RedisProjectCache) GetProject(ctx context.Context, projectID int64) (dto.ProjectResponse, bool, error) {
	var project dto.ProjectResponse

	data, err := c.client.Get(ctx, projectKey(projectID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return dto.ProjectResponse{}, false, nil
		}
		return dto.ProjectResponse{}, false, err
	}

	if err := json.Unmarshal([]byte(data), &project); err != nil {
		_ = c.DeleteProject(ctx, projectID) //cache miss
		return dto.ProjectResponse{}, false, err
	}
	return project, true, nil
}

func (c *RedisProjectCache) SetProject(ctx context.Context, project dto.ProjectResponse) error {
	if project.ID <= 0 {
		return errors.New("invalid project id")
	}

	data, err := json.Marshal(project)
	if err != nil {
		return err
	}

	return c.client.Set(
		ctx,
		projectKey(project.ID),
		data,
		c.ttl,
	).Err()
}

func (c *RedisProjectCache) DeleteProject(ctx context.Context, projectID int64) error {
	return c.client.Del(ctx, projectKey(projectID)).Err()
}
