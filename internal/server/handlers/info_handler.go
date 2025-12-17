package handlers

import (
	"net/http"
)

type InfoHandler struct {
	BuildVersion string `json:"version"`
	BuildDate    string `json:"build"`
}

func NewInfoHandler(version, build string) *InfoHandler {
	return &InfoHandler{
		BuildVersion: version,
		BuildDate:    build,
	}
}

func (ih *InfoHandler) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (ih *InfoHandler) Version(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, ih)
}
