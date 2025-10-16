package repository

import (
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

// Repositories aggregates low-level data dependencies for DI.
type Repositories struct {
	MongoDB *mongo.Database
	Redis   *redis.Client
}

// New creates a new Repositories struct bound to provided stores.
func New(db *mongo.Database, redisClient *redis.Client) *Repositories {
	return &Repositories{
		MongoDB: db,
		Redis:   redisClient,
	}
}
