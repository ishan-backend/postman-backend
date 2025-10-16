package service

import (
	"context"

	"github.com/ishan-backend/postman-backend/repository"
)

// Services aggregates all business logic services.
type Services struct {
	repos *repository.Repositories
	Users *UserService
}

// New creates Services bound to repositories.
func New(repos *repository.Repositories) *Services {
	return &Services{repos: repos, Users: NewUserService(repos)}
}

// RedisPing checks connectivity to Redis via repository Redis client.
func (s *Services) RedisPing(ctx context.Context) error {
	if s == nil || s.repos == nil || s.repos.Redis == nil {
		return nil
	}
	return s.repos.Redis.Ping(ctx).Err()
}

// MongoListCollections pings Mongo by listing all collection names.
// Returns the collection names if successful; error if Mongo is unavailable.
func (s *Services) MongoListCollections(ctx context.Context) ([]string, error) {
	if s == nil || s.repos == nil || s.repos.MongoDB == nil {
		return nil, nil
	}
	collections, err := s.repos.MongoDB.ListCollectionNames(ctx, struct{}{})
	if err != nil {
		return nil, err
	}
	return collections, nil
}
