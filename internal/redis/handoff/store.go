package handoff

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrCodeNotFound = errors.New("handoff code not found")

type Store struct {
	client *redis.Client
}

func NewStore(client *redis.Client) *Store {
	return &Store{
		client: client,
	}
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (s *Store) Save(ctx context.Context, code string, t Tokens, ttl time.Duration) error {
	key := "handoff:" + code

	//store as two fields, to return in a single round trip
	err := s.client.HSet(ctx, key,
		"access_token", t.AccessToken,
		"refresh_token", t.RefreshToken,
	).Err()

	if err != nil {
		return fmt.Errorf("hset: %w", err)
	}
	if err := s.client.Expire(ctx, key, ttl).Err(); err != nil {
		return fmt.Errorf("expire: %w", err)
	}
	return nil
}

func (s *Store) Consume(ctx context.Context, code string) (Tokens, error) {
	key := "handoff:" + code

	vals, err := s.client.HGetAll(ctx, key).Result()
	if err != nil {
		return Tokens{}, fmt.Errorf("hgetall: %w", err)
	}
	if len(vals) == 0 {
		return Tokens{}, ErrCodeNotFound
	}

	//delete before returning
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return Tokens{}, fmt.Errorf("del: %w", err)
	}

	return Tokens{
		AccessToken:  vals["access_token"],
		RefreshToken: vals["refresh_token"],
	}, nil
}
