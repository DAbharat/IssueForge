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

type WorkspaceInvitationCache interface {
	GetPendingForUser(ctx context.Context, userID int64) ([]dto.PendingWorkspaceInvitationResponse, bool, error)
	SetPendingForUser(ctx context.Context, userID int64, invitations []dto.PendingWorkspaceInvitationResponse) error
	DeletePendingForUser(ctx context.Context, userID int64) error
	GetPendingForWorkspace(ctx context.Context, workspaceID int64) ([]dto.WorkspacePendingInvitationResponse, bool, error)
	SetPendingForWorkspace(ctx context.Context, workspaceID int64, invitations []dto.WorkspacePendingInvitationResponse) error
	DeletePendingForWorkspace(ctx context.Context, workspaceID int64) error
}

type RedisWorkspaceInvitationCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewWorkspaceInvitationCache(client *redis.Client, ttl time.Duration) *RedisWorkspaceInvitationCache {
	if ttl == 0 {
		ttl = 10 * time.Minute
	}
	return &RedisWorkspaceInvitationCache{
		client: client,
		ttl:    ttl,
	}
}

func userInvitationKey(userID int64) string {
	return fmt.Sprintf("workspace-invitations:user:%d", userID)
}
func workspaceInvitationKey(workspaceID int64) string {
	return fmt.Sprintf("workspace-invitations-workspace:%d", workspaceID)
}

func (c *RedisWorkspaceInvitationCache) GetPendingForUser(ctx context.Context, userID int64) ([]dto.PendingWorkspaceInvitationResponse, bool, error) {
	var user []dto.PendingWorkspaceInvitationResponse

	data, err := c.client.Get(ctx, userInvitationKey(userID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil
		}
		return nil, false, err
	}

	if err := json.Unmarshal([]byte(data), &user); err != nil {
		_ = c.DeletePendingForUser(ctx, userID)
		return nil, false, err
	}
	return user, true, nil
}

func (c *RedisWorkspaceInvitationCache) SetPendingForUser(ctx context.Context, userID int64, invitation []dto.PendingWorkspaceInvitationResponse) error {
	if userID <= 0 {
		return errors.New("invalid user id")
	}

	data, err := json.Marshal(invitation)
	if err != nil {
		return err
	}

	return c.client.Set(
		ctx,
		userInvitationKey(userID),
		data,
		c.ttl,
	).Err()
}

func (c *RedisWorkspaceInvitationCache) DeletePendingForUser(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return errors.New("invalid user id")
	}

	return c.client.Del(ctx, userInvitationKey(userID)).Err()
}

func (c *RedisWorkspaceInvitationCache) GetPendingForWorkspace(ctx context.Context, workspaceID int64) ([]dto.WorkspacePendingInvitationResponse, bool, error) {
	var workspace []dto.WorkspacePendingInvitationResponse

	data, err := c.client.Get(ctx, workspaceInvitationKey(workspaceID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil
		}
		return nil, false, err
	}

	if err := json.Unmarshal([]byte(data), &workspace); err != nil {
		_ = c.DeletePendingForWorkspace(ctx, workspaceID)
		return nil, false, err
	}
	return workspace, true, nil
}

func (c *RedisWorkspaceInvitationCache) SetPendingForWorkspace(ctx context.Context, workspaceID int64, invitation []dto.WorkspacePendingInvitationResponse) error {
	if workspaceID <= 0 {
		return errors.New("invalid workspace id")
	}

	data, err := json.Marshal(invitation)
	if err != nil {
		return err
	}

	return c.client.Set(
		ctx,
		workspaceInvitationKey(workspaceID),
		data,
		c.ttl,
	).Err()
}

func (c *RedisWorkspaceInvitationCache) DeletePendingForWorkspace(ctx context.Context, workspaceID int64) error {
	if workspaceID <= 0 {
		return errors.New("invalid workspace id")
	}

	return c.client.Del(ctx, workspaceInvitationKey(workspaceID)).Err()
}
