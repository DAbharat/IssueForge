package middleware

import (
	"IssueForge/internal/httpx"
	"IssueForge/internal/redis"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
)

type RateLimitMiddleware struct {
	limiter    *redis.RateLimiter
	capacity   int
	refillRate float64
}

func NewRateLimitMiddleware(limiter *redis.RateLimiter, capacity int, refillRate float64) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter:    limiter,
		capacity:   capacity,
		refillRate: refillRate,
	}
}

func (m *RateLimitMiddleware) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var key string

		userID, ok := GetUserFromContext(r.Context())
		if ok {
			key = fmt.Sprintf("rate_limit:user:%d", userID)
		} else {
			clientIP := extractIP(r)
			key = fmt.Sprintf("rate_limit:ip:%s", clientIP)
		}

		result, err := m.limiter.Allow(r.Context(), key, m.capacity, m.refillRate, 1)
		if err != nil {
			httpx.RespondWithError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(m.capacity))

		if !result.Allowed {
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("Retry-After", "1")
			httpx.RespondWithError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}

		w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(result.RemainingTokens, 10))

		next.ServeHTTP(w, r)
	})
}

func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		clientIP := strings.TrimSpace(ips[0])
		if clientIP != "" {
			return clientIP
		}
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}
