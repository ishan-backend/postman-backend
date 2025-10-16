package db

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis wraps a connected Redis client.
type Redis struct {
	Client *redis.Client
}

var redisInstance *Redis

// InitRedis initializes a singleton Redis client using provided config.
// Safe to call multiple times; the first successful call wins.
func InitRedis(addr string, password string, db int, dialTimeoutSeconds int, readTimeoutSeconds int, writeTimeoutSeconds int) (*Redis, error) {
	if redisInstance != nil {
		return redisInstance, nil
	}

	if dialTimeoutSeconds <= 0 {
		dialTimeoutSeconds = 5
	}
	if readTimeoutSeconds <= 0 {
		readTimeoutSeconds = 3
	}
	if writeTimeoutSeconds <= 0 {
		writeTimeoutSeconds = 3
	}

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  time.Duration(dialTimeoutSeconds) * time.Second,
		ReadTimeout:  time.Duration(readTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(writeTimeoutSeconds) * time.Second,
	})

	// Ping to verify connectivity
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	redisInstance = &Redis{Client: client}
	return redisInstance, nil
}

// GetRedis returns the initialized Redis singleton, or nil if not initialized.
func GetRedis() *Redis {
	return redisInstance
}

// Close terminates the underlying Redis client.
func (r *Redis) Close() error {
	if r == nil || r.Client == nil {
		return nil
	}
	return r.Client.Close()
}
