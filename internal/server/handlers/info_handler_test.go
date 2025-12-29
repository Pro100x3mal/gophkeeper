package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInfoHandler(t *testing.T) {
	version := "1.0.0"
	build := "2024-01-01"

	handler := NewInfoHandler(version, build)

	assert.NotNil(t, handler)
	assert.Equal(t, version, handler.BuildVersion)
	assert.Equal(t, build, handler.BuildDate)
}

func TestInfoHandler_HealthCheck(t *testing.T) {
	handler := NewInfoHandler("1.0.0", "2024-01-01")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.HealthCheck(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

func TestInfoHandler_Version(t *testing.T) {
	version := "2.5.1"
	build := "2024-12-25"
	handler := NewInfoHandler(version, build)

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	w := httptest.NewRecorder()

	handler.Version(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response InfoHandler
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, version, response.BuildVersion)
	assert.Equal(t, build, response.BuildDate)
}

func TestInfoHandler_MultipleHealthChecks(t *testing.T) {
	handler := NewInfoHandler("1.0.0", "2024-01-01")

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		handler.HealthCheck(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}
}
