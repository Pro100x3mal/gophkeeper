// Package middleware provides HTTP middleware functions for the GophKeeper server.
//
// This package implements authentication, logging, and request context management middleware.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Pro100x3mal/gophkeeper/pkg/jwt"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type contextKey string

const userIDContextKey contextKey = "user_id"

// Auth returns a middleware that validates JWT tokens from the Authorization header.
// Extracts the user ID from valid tokens and adds it to the request context.
func Auth(jwtGen *jwt.Generator, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpLog := logger.With(zap.String("middleware", "auth"))

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			userID, err := jwtGen.ValidateToken(parts[1])
			if err != nil {
				httpLog.Error("invalid jwt token", zap.Error(err))
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserIDFromContext extracts the user ID from the request context.
// Returns the user ID and true if found, or uuid.Nil and false otherwise.
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	val := ctx.Value(userIDContextKey)
	if val == nil {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}
