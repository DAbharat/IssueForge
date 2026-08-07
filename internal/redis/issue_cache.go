package redis

import (
	"IssueForge/internal/dto"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type IssueCache interface {
	GetIssue(ctx context.Context, issueID int64) (dto.IssueResponse, bool, error)
	SetIssue(ctx context.Context, issue dto.IssueResponse) error
	DeleteIssue(ctx context.Context, issueID int64) error
}

type RedisIssueCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewIssueCache(client *redis.Client, ttl time.Duration) *RedisIssueCache {
	if ttl == 0 {
		ttl = 10 * time.Minute
	}
	return &RedisIssueCache{
		client: client,
		ttl:    ttl,
	}
}

func issueKey(issueID int64) string {
	return fmt.Sprintf("issue:%d", issueID)
}

func (c *RedisIssueCache) GetIssue(ctx context.Context, issueID int64) (dto.IssueResponse, bool, error) {
	var issue dto.IssueResponse

	data, err := c.client.Get(ctx, issueKey(issueID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return dto.IssueResponse{}, false, nil
		}
		return dto.IssueResponse{}, false, err
	}

	if err := json.Unmarshal([]byte(data), &issue); err != nil {
		return dto.IssueResponse{}, false, err
	}
	return issue, true, nil
}

func (c *RedisIssueCache) SetIssue(ctx context.Context, issue dto.IssueResponse) error {
	if issue.ID <= 0 {
		return errors.New("invalid issue id")
	}

	data, err := json.Marshal(issue)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, issueKey(issue.ID), data, c.ttl).Err()
}

func (c *RedisIssueCache) DeleteIssue(ctx context.Context, issueID int64) error {
	return c.client.Del(ctx, issueKey(issueID)).Err()
}
