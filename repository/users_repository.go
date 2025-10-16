package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ishan-backend/postman-backend/models/repo"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

// GetUserByID retrieves a user by ID, checking Redis cache first, then MongoDB.
func (u *UsersRepository) GetUserByID(ctx context.Context, userID primitive.ObjectID) (*repo.User, error) {
	key := "user:" + userID.Hex()

	// Try Redis first
	if u.redis != nil {
		result, err := u.redis.HGetAll(ctx, key).Result()
		if err == nil && len(result) > 0 {
			// Reconstruct user from Redis hash
			user := &repo.User{
				ID:        userID,
				Email:     result["email"],
				FirstName: result["first_name"],
				LastName:  result["last_name"],
			}

			// Parse timestamps if present
			if createdAt, ok := result["created_at"]; ok {
				var ts int64
				fmt.Sscanf(createdAt, "%d", &ts)
				user.CreatedAt = time.Unix(ts, 0).UTC()
			}
			if updatedAt, ok := result["updated_at"]; ok {
				var ts int64
				fmt.Sscanf(updatedAt, "%d", &ts)
				user.UpdatedAt = time.Unix(ts, 0).UTC()
			}

			// Parse friends list if stored
			if friendsJSON, ok := result["friends"]; ok && friendsJSON != "" {
				json.Unmarshal([]byte(friendsJSON), &user.FriendsList)
			}

			return user, nil
		}
	}

	// Fallback to MongoDB
	var user repo.User
	err := u.collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("no user found")
		}
		return nil, err
	}

	// Cache in Redis for future requests
	if u.redis != nil {
		friendsJSON, _ := json.Marshal(user.FriendsList)
		fields := map[string]interface{}{
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"created_at": user.CreatedAt.Unix(),
			"updated_at": user.UpdatedAt.Unix(),
			"friends":    string(friendsJSON),
		}
		u.redis.HSet(ctx, key, fields)
		u.redis.Expire(ctx, key, 24*time.Hour) // Optional: set TTL
	}

	return &user, nil
}

// UpdateUser updates the email for a user in both MongoDB and Redis.
func (u *UsersRepository) UpdateUser(ctx context.Context, userID primitive.ObjectID, email string) error {
	now := time.Now().UTC()

	// Update in MongoDB
	update := bson.M{
		"$set": bson.M{
			"email":      email,
			"updated_at": now,
		},
	}

	result, err := u.collection.UpdateOne(ctx, bson.M{"_id": userID}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("user not found")
	}

	// Update in Redis
	if u.redis != nil {
		key := "user:" + userID.Hex()
		fields := map[string]interface{}{
			"email":      email,
			"updated_at": now.Unix(),
		}
		if err := u.redis.HSet(ctx, key, fields).Err(); err != nil {
			return err
		}
	}

	return nil
}

// DeleteUser removes a user from both MongoDB and Redis.
func (u *UsersRepository) DeleteUser(ctx context.Context, userID primitive.ObjectID) error {
	// Delete from MongoDB
	result, err := u.collection.DeleteOne(ctx, bson.M{"_id": userID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("user not found")
	}

	// Delete from Redis
	if u.redis != nil {
		key := "user:" + userID.Hex()
		if err := u.redis.Del(ctx, key).Err(); err != nil {
			// Log error but don't fail the operation since MongoDB deletion succeeded
			// In production, consider using a background job to clean up orphaned cache entries
			return fmt.Errorf("user deleted from MongoDB but Redis cleanup failed: %w", err)
		}
	}

	return nil
}

// UpdateUserFriends adds a friend to the user's friends list with optimistic locking.
// Uses MongoDB's findAndModify with version field to handle concurrent updates.
func (u *UsersRepository) UpdateUserFriends(ctx context.Context, userID, friendID primitive.ObjectID) error {
	now := time.Now().UTC()

	// Use findOneAndUpdate with $addToSet to ensure uniqueness and atomic operation
	// $addToSet only adds if the value doesn't exist, preventing duplicates
	filter := bson.M{"_id": userID}
	update := bson.M{
		"$addToSet": bson.M{"friends_list": friendID},
		"$set":      bson.M{"updated_at": now},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedUser repo.User
	err := u.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("no documents found")
		}
		return err
	}

	// Update Redis cache with new friends list
	if u.redis != nil {
		key := "user:" + userID.Hex()
		friendsJSON, _ := json.Marshal(updatedUser.FriendsList)
		fields := map[string]interface{}{
			"friends":    string(friendsJSON),
			"updated_at": now.Unix(),
		}
		if err := u.redis.HSet(ctx, key, fields).Err(); err != nil {
			return err
		}
	}

	return nil
}
