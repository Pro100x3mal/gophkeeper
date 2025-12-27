// Package handlers provides HTTP request handlers for the GophKeeper server API.
//
// This package implements handlers for authentication, item management, and system information endpoints.
package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

// isJSON checks if the request Content-Type header indicates JSON.
func isJSON(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Content-Type"), "application/json")
}

// writeJSON writes a JSON response with the specified status code.
// Sets the Content-Type header and encodes the provided value as JSON.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
