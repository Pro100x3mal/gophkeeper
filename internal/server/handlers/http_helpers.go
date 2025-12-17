package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

func isJSON(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Content-Type"), "application/json")
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
