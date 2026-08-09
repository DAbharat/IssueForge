package refreshtoken

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store struct {
	client *redis.Client
	ttl    time.Duration
}

func NewStore(client *redis.Client, ttl time.Duration) *Store {
	return &Store{
		client: client,
		ttl:    ttl,
	}
}

func (s *Store) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (s *Store) key(token string) string {
	return fmt.Sprintf("refresh_token:%s", s.hashToken(token))
}

var ErrTokenNotFound = errors.New("refresh token expired or does not exist")

func (s *Store) CreateRefreshToken(ctx context.Context, token string, userID int64) error {
	err := s.client.Set(ctx, s.key(token), userID, s.ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	return nil
}

func (s *Store) GetRefreshToken(ctx context.Context, token string) (int64, error) {
	val, err := s.client.Get(ctx, s.key(token)).Result()
	if errors.Is(err, redis.Nil) {
		return 0, ErrTokenNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get refresh token: %w", err)
	}

	userID, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid user id in refresh token storage: %w", err)
	}

	return userID, nil
}

func (s *Store) DeleteRefreshToken(ctx context.Context, token string) error {
	err := s.client.Del(ctx, s.key(token)).Err()
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return nil
}
