package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"github.com/ishan-backend/postman-backend/config"
	"github.com/ishan-backend/postman-backend/db"
	"github.com/ishan-backend/postman-backend/handlers"
	"github.com/ishan-backend/postman-backend/repository"
	"github.com/ishan-backend/postman-backend/service"
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

	// Initialize Mongo and repositories
	mongo, err := db.InitMongo(
		cfg.Mongo.URI,
		cfg.Mongo.Database,
		cfg.Mongo.ConnectTimeout,
		cfg.Mongo.Username,
		cfg.Mongo.Password,
		cfg.Mongo.AuthSource,
	)
	if err != nil {
		log.Fatalf("failed to initialize mongo: %v", err)
	}

	// Initialize Redis
	_, err = db.InitRedis(
		cfg.Redis.Addr,
		cfg.Redis.Password,
		cfg.Redis.DB,
		cfg.Redis.DialTimeout,
		cfg.Redis.ReadTimeout,
		cfg.Redis.WriteTimeout,
	)
	if err != nil {
		log.Fatalf("failed to initialize redis: %v", err)
	}

	repos := repository.New(mongo.Database, db.GetRedis().Client)
	services := service.New(repos)
	api := handlers.New(services)

	// Healthcheck endpoints
	r.HandleFunc("/ping", api.Ping).Methods(http.MethodGet)
	r.HandleFunc("/health", api.Health).Methods(http.MethodGet)
	r.HandleFunc("/redis-ping", api.RedisPing).Methods(http.MethodGet)
	r.HandleFunc("/mongo-ping", api.MongoPing).Methods(http.MethodGet)

	// Users
	usersRouter := r.PathPrefix("/users").Subrouter()
	usersRouter.Use(authMiddleware)
	usersRouter.HandleFunc("/bulk", api.BulkCreateUsers).Methods(http.MethodPost)

	addr := cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port)
	log.Printf("starting server on %s", addr)

	// Apply request timeout if configured
	var rootHandler http.Handler = r
	if cfg.Server.RequestTimeoutSeconds > 0 {
		rootHandler = http.TimeoutHandler(r, time.Duration(cfg.Server.RequestTimeoutSeconds)*time.Second, "request timeout")
	}
	// Graceful shutdown handling (for DB cleanup)
	shutdownCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srvErr := make(chan error, 1)
	go func() {
		srvErr <- http.ListenAndServe(addr, rootHandler)
	}()

	select {
	case <-shutdownCtx.Done():
		// attempt to close DB
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if mongo != nil {
			_ = mongo.Close(ctx)
		}
		// attempt to close Redis
		if r := db.GetRedis(); r != nil {
			_ = r.Close()
		}
		log.Printf("shutting down")
		os.Exit(0)
	case err := <-srvErr:
		if err != nil {
			log.Fatalf("server error: %v", err)
		}
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
