package middleware

import (
	"IssueForge/internal/auth"
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
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := auth.ParseToken(tokenString, m.jwtSecret)
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrExpiredToken):
				http.Error(w, "token expired", http.StatusUnauthorized)
			default:
				http.Error(w, "invalid token", http.StatusUnauthorized)
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
