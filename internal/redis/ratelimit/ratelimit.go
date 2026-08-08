package ratelimit

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed scripts/token_bucket.lua
var tokenBucket string

type RateLimiter struct {
	client   *redis.Client
	script   *redis.Script
	failOpen bool
}

type RateLimitResult struct {
	Allowed         bool
	RemainingTokens int64
}

func NewRateLimiter(client *redis.Client, failOpen bool) *RateLimiter {
	return &RateLimiter{
		client:   client,
		script:   redis.NewScript(tokenBucket),
		failOpen: failOpen,
	}
}

func (r *RateLimiter) Allow(ctx context.Context, key string, capacity int, refillRate float64, requested int) (RateLimitResult, error) {
	evalCtx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()

	now := float64(time.Now().UnixNano()) / 1e9

	result, err := r.script.Run(evalCtx, r.client, []string{key}, capacity, refillRate, now, requested).Result()
	if err != nil {
		if r.failOpen {
			log.Printf("RateLimiter Redis failure (failing open): %v", err)
			return RateLimitResult{
				Allowed:         true,
				RemainingTokens: 0,
			}, nil
		}
		return RateLimitResult{}, fmt.Errorf("rate limiter redis execution failed: %w", err)
	}

	values, ok := result.([]interface{})
	if !ok || len(values) != 2 {
		return RateLimitResult{}, fmt.Errorf("unexpected lua result: %v", result)
	}

	allowed, ok := values[0].(int64)
	if !ok {
		return RateLimitResult{}, fmt.Errorf("invalid allowed value")
	}

	remaining, ok := values[1].(int64)
	if !ok {
		return RateLimitResult{}, fmt.Errorf("invalid remaining tokens value")
	}

	return RateLimitResult{
		Allowed:         allowed == 1,
		RemainingTokens: remaining,
	}, nil
}
