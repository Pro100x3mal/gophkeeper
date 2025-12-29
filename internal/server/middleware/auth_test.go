package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Pro100x3mal/gophkeeper/pkg/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAuth_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	jwtGen := jwt.NewGenerator("secret", time.Hour)
	userID := uuid.New()

	token, err := jwtGen.GenerateToken(userID)
	require.NoError(t, err)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		extractedUserID, ok := GetUserIDFromContext(r.Context())
		assert.True(t, ok)
		assert.Equal(t, userID, extractedUserID)
		w.WriteHeader(http.StatusOK)
	})

	middleware := Auth(jwtGen, logger)
	handler := middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuth_MissingAuthHeader(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	jwtGen := jwt.NewGenerator("secret", time.Hour)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	middleware := Auth(jwtGen, logger)
	handler := middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_InvalidAuthHeaderFormat(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	jwtGen := jwt.NewGenerator("secret", time.Hour)

	tests := []struct {
		name   string
		header string
	}{
		{"No Bearer prefix", "token"},
		{"Wrong prefix", "Basic token"},
		{"Extra parts", "Bearer token extra"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
			})

			middleware := Auth(jwtGen, logger)
			handler := middleware(next)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", tt.header)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.False(t, nextCalled)
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	jwtGen := jwt.NewGenerator("secret", time.Hour)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	middleware := Auth(jwtGen, logger)
	handler := middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_ExpiredToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	jwtGen := jwt.NewGenerator("secret", 1*time.Millisecond)
	userID := uuid.New()

	token, err := jwtGen.GenerateToken(userID)
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	middleware := Auth(jwtGen, logger)
	handler := middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_CaseInsensitiveBearer(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	jwtGen := jwt.NewGenerator("secret", time.Hour)
	userID := uuid.New()

	token, err := jwtGen.GenerateToken(userID)
	require.NoError(t, err)

	tests := []struct {
		name   string
		prefix string
	}{
		{"lowercase bearer", "bearer"},
		{"uppercase BEARER", "BEARER"},
		{"mixed case Bearer", "Bearer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			middleware := Auth(jwtGen, logger)
			handler := middleware(next)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", tt.prefix+" "+token)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.True(t, nextCalled)
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestGetUserIDFromContext_Success(t *testing.T) {
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), userIDContextKey, userID)

	extractedID, ok := GetUserIDFromContext(ctx)

	assert.True(t, ok)
	assert.Equal(t, userID, extractedID)
}

func TestGetUserIDFromContext_NotFound(t *testing.T) {
	ctx := context.Background()

	extractedID, ok := GetUserIDFromContext(ctx)

	assert.False(t, ok)
	assert.Equal(t, uuid.Nil, extractedID)
}

func TestGetUserIDFromContext_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), userIDContextKey, "not-a-uuid")

	extractedID, ok := GetUserIDFromContext(ctx)

	assert.False(t, ok)
	assert.Equal(t, uuid.Nil, extractedID)
}
