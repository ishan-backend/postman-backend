package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ishan-backend/postman-backend/service"
)

// API wires HTTP handlers with repositories and other dependencies.
type API struct {
	Services *service.Services
}

// New creates a new API instance.
func New(services *service.Services) *API {
	return &API{Services: services}
}

// Ping responds with a simple OK status and body for quick health checks.
func (a *API) Ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// Health responds with OK status and body for liveness/readiness probes.
func (a *API) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// RedisPing checks Redis connectivity via service.
func (a *API) RedisPing(w http.ResponseWriter, r *http.Request) {
	if err := a.Services.RedisPing(r.Context()); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("redis unavailable"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("pong"))
}

// MongoPing lists all MongoDB collection names to verify connectivity.
func (a *API) MongoPing(w http.ResponseWriter, r *http.Request) {
	names, err := a.Services.MongoListCollections(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("mongo unavailable"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"collections": names,
	})
}
