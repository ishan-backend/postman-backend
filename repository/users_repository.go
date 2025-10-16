package repository

import (
	"context"
	"errors"
	"time"

	"github.com/ishan-backend/postman-backend/models/repo"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// UsersRepoInterface defines contract for interacting with users store.
type UsersRepoInterface interface {
	BulkInsert(ctx context.Context, users []repo.User) ([]primitive.ObjectID, error)
}

// UsersRepository is a concrete repository with Mongo and Redis dependencies.
type UsersRepository struct {
	db         *mongo.Database
	redis      *redis.Client
	collection *mongo.Collection
}

// NewUsersRepository constructs a UsersRepository bound to the given Repositories.
func NewUsersRepository(repos *Repositories) *UsersRepository {
	return &UsersRepository{
		db:         repos.MongoDB,
		redis:      repos.Redis,
		collection: repos.MongoDB.Collection("users"),
	}
}

// BulkInsert inserts users and mirrors data in Redis via TxPipelined.
func (u *UsersRepository) BulkInsert(ctx context.Context, users []repo.User) ([]primitive.ObjectID, error) {
	if len(users) == 0 {
		return nil, nil
	}

	// Ensure timestamps and ObjectIDs
	now := time.Now().UTC()
	docs := make([]interface{}, 0, len(users))
	ids := make([]primitive.ObjectID, 0, len(users))
	for i := range users {
		if users[i].ID.IsZero() {
			users[i].ID = primitive.NewObjectID()
		}
		users[i].CreatedAt = now
		users[i].UpdatedAt = now
		ids = append(ids, users[i].ID)
		docs = append(docs, users[i])
	}

	// Insert many directly without MongoDB transactions (works on standalone server)
	if _, err := u.collection.InsertMany(ctx, docs); err != nil {
		return nil, err
	}

	// Mirror to Redis atomically using TxPipelined
	if u.redis != nil {
		if _, err := u.redis.TxPipelined(ctx, func(p redis.Pipeliner) error {
			for _, usr := range users {
				key := "user:" + usr.ID.Hex()
				fields := map[string]interface{}{
					"email":      usr.Email,
					"first_name": usr.FirstName,
					"last_name":  usr.LastName,
					"created_at": usr.CreatedAt.Unix(),
					"updated_at": usr.UpdatedAt.Unix(),
				}
				p.HSet(ctx, key, fields)
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}

	return ids, nil
}

// Ensure indexes for users (e.g., unique email). Optional utility for setup.
func (u *UsersRepository) EnsureIndexes(ctx context.Context) error {
	if u.collection == nil {
		return errors.New("users collection not initialized")
	}
	// Unique index on email
	_, err := u.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: nil, // could set unique via options.Index().SetUnique(true)
	})
	return err
}
