package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRequireUser_Success(t *testing.T) {
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), userIDContextKey, userID)

	handlerCalled := false
	handler := RequireUser(func(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
		handlerCalled = true
		assert.Equal(t, userID, id)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireUser_NoUserInContext(t *testing.T) {
	ctx := context.Background()

	handlerCalled := false
	handler := RequireUser(func(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
		handlerCalled = true
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireUser_WrongTypeInContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), userIDContextKey, "not-a-uuid")

	handlerCalled := false
	handler := RequireUser(func(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
		handlerCalled = true
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireUser_PreservesRequest(t *testing.T) {
	userID := uuid.New()
	ctx := context.WithValue(context.Background(), userIDContextKey, userID)

	handler := RequireUser(func(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/test/path", r.URL.Path)
		assert.Equal(t, "test-value", r.Header.Get("X-Test-Header"))
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/test/path", nil)
	req.Header.Set("X-Test-Header", "test-value")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireUser_DifferentUserIDs(t *testing.T) {
	userIDs := []uuid.UUID{
		uuid.New(),
		uuid.New(),
		uuid.New(),
	}

	for _, userID := range userIDs {
		ctx := context.WithValue(context.Background(), userIDContextKey, userID)

		handler := RequireUser(func(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
			assert.Equal(t, userID, id)
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}
}
