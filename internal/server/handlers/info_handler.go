package handlers

import (
	"net/http"
)

// InfoHandler handles system information and health check endpoints.
type InfoHandler struct {
	BuildVersion string `json:"version"`
	BuildDate    string `json:"build"`
}

// NewInfoHandler creates a new info handler with build information.
func NewInfoHandler(version, build string) *InfoHandler {
	return &InfoHandler{
		BuildVersion: version,
		BuildDate:    build,
	}
}

// HealthCheck handles health check requests returning the service status.
func (ih *InfoHandler) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Version handles version information requests returning build details.
func (ih *InfoHandler) Version(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, ih)
}
