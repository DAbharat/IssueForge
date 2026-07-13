package middleware

import (
	"IssueForge/internal/auth"
	"IssueForge/internal/httpx"
	"context"
	"errors"
	"net/http"
	"strings"
)

type contextKey string

const userIDKey contextKey = "userID"

type AuthMiddleware struct {
	jwtSecret string
}

func NewAuthMiddleware(secret string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: secret,
	}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			httpx.RespondWithError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		parts := strings.Fields(authHeader)
		if len(parts) < 2 || parts[0] != "Bearer" {
			httpx.RespondWithError(w, http.StatusUnauthorized, "invalid authorization header")
			return
		}

		tokenString := parts[1]

		claims, err := auth.ParseToken(tokenString, m.jwtSecret)
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrExpiredToken):
				httpx.RespondWithError(w, http.StatusUnauthorized, "token expired")
			default:
				httpx.RespondWithError(w, http.StatusUnauthorized, "invalid token")
			}
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDKey).(int64)
	if !ok {
		return 0, false
	}
	return userID, true
}
