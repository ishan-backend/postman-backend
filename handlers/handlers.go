package handlers

import (
	"net/http"
)

// Ping responds with a simple OK status and body for quick health checks.
func Ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// Health responds with OK status and body for liveness/readiness probes.
func Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
