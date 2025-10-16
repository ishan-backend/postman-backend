package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/ishan-backend/postman-backend/config"
	"github.com/ishan-backend/postman-backend/handlers"
)

func main() {
	// Load configuration (from CONFIG_PATH or ./config.yaml)
	config.MustLoad("./config.yaml")
	cfg, err := config.Get()
	if err != nil {
		log.Fatalf("failed to get config: %v", err)
	}

	// Global middlewares
	r := mux.NewRouter()
	r.Use(loggingMiddleware)

	// Healthcheck endpoints
	r.HandleFunc("/ping", handlers.Ping).Methods(http.MethodGet)
	r.HandleFunc("/health", handlers.Health).Methods(http.MethodGet)

	addr := cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port)
	log.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// loggingMiddleware logs method, path, and request duration for each request.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
