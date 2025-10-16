package main

import (
	"log"
	"net/http"
	"strconv"

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

	r := mux.NewRouter()

	// Healthcheck endpoints
	r.HandleFunc("/ping", handlers.Ping).Methods(http.MethodGet)
	r.HandleFunc("/health", handlers.Health).Methods(http.MethodGet)

	addr := cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port)
	log.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
