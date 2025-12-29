package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsJSON_True(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
	}{
		{"Simple JSON", "application/json"},
		{"JSON with charset", "application/json; charset=utf-8"},
		{"JSON with spaces", "application/json "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			req.Header.Set("Content-Type", tt.contentType)

			result := isJSON(req)

			assert.True(t, result)
		})
	}
}

func TestIsJSON_False(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
	}{
		{"Empty", ""},
		{"Plain text", "text/plain"},
		{"XML", "application/xml"},
		{"Form data", "application/x-www-form-urlencoded"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			req.Header.Set("Content-Type", tt.contentType)

			result := isJSON(req)

			assert.False(t, result)
		})
	}
}

func TestWriteJSON_Success(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}

	writeJSON(w, http.StatusOK, data)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), `"key":"value"`)
}

func TestWriteJSON_DifferentStatuses(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{"OK", http.StatusOK},
		{"Created", http.StatusCreated},
		{"BadRequest", http.StatusBadRequest},
		{"Unauthorized", http.StatusUnauthorized},
		{"NotFound", http.StatusNotFound},
		{"InternalServerError", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			data := map[string]string{"status": tt.name}

			writeJSON(w, tt.status, data)

			assert.Equal(t, tt.status, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		})
	}
}

func TestWriteJSON_ComplexData(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]interface{}{
		"string":  "value",
		"number":  123,
		"boolean": true,
		"nested": map[string]string{
			"inner": "data",
		},
		"array": []int{1, 2, 3},
	}

	writeJSON(w, http.StatusOK, data)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"string":"value"`)
	assert.Contains(t, w.Body.String(), `"number":123`)
	assert.Contains(t, w.Body.String(), `"boolean":true`)
}
