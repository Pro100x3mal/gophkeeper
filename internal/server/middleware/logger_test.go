package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestLogger_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	middleware := Logger(logger)
	handler := middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test response", w.Body.String())
}

func TestLogger_CapturesStatus(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	statuses := []int{
		http.StatusOK,
		http.StatusCreated,
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusNotFound,
		http.StatusInternalServerError,
	}

	for _, expectedStatus := range statuses {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(expectedStatus)
		})

		middleware := Logger(logger)
		handler := middleware(next)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, expectedStatus, w.Code)
	}
}

func TestLogger_CapturesSize(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	testCases := []struct {
		name         string
		response     string
		expectedSize int
	}{
		{"Empty response", "", 0},
		{"Small response", "test", 4},
		{"Larger response", "this is a longer test response", 30},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(tc.response))
			})

			middleware := Logger(logger)
			handler := middleware(next)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.response, w.Body.String())
			assert.Equal(t, tc.expectedSize, w.Body.Len())
		})
	}
}

func TestLogger_DefaultStatusOK(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't explicitly set status
		w.Write([]byte("response"))
	})

	middleware := Logger(logger)
	handler := middleware(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWrappedResponseWriter_Write(t *testing.T) {
	w := httptest.NewRecorder()
	wrw := &wrappedResponseWriter{
		ResponseWriter: w,
	}

	data := []byte("test data")
	n, err := wrw.Write(data)

	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, len(data), wrw.size)
	assert.Equal(t, http.StatusOK, wrw.status)
}

func TestWrappedResponseWriter_WriteHeader(t *testing.T) {
	w := httptest.NewRecorder()
	wrw := &wrappedResponseWriter{
		ResponseWriter: w,
	}

	wrw.WriteHeader(http.StatusCreated)

	assert.Equal(t, http.StatusCreated, wrw.status)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestWrappedResponseWriter_MultipleWrites(t *testing.T) {
	w := httptest.NewRecorder()
	wrw := &wrappedResponseWriter{
		ResponseWriter: w,
	}

	wrw.Write([]byte("first "))
	wrw.Write([]byte("second "))
	wrw.Write([]byte("third"))

	assert.Equal(t, 18, wrw.size) // "first second third" = 18 bytes
	assert.Equal(t, "first second third", w.Body.String())
}

func TestLogger_DifferentHTTPMethods(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
	}

	for _, method := range methods {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, method, r.Method)
			w.WriteHeader(http.StatusOK)
		})

		middleware := Logger(logger)
		handler := middleware(next)

		req := httptest.NewRequest(method, "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}
}
