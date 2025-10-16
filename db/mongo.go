package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Mongo wraps a connected mongo client and database handle.
type Mongo struct {
	Client   *mongo.Client
	Database *mongo.Database
}

var mongoInstance *Mongo

// InitMongo initializes a singleton Mongo connection using provided config.
// Safe to call multiple times; the first successful call wins.
func InitMongo(uri string, database string, connectTimeoutSeconds int, username string, password string, authSource string) (*Mongo, error) {
	if mongoInstance != nil {
		return mongoInstance, nil
	}

	if connectTimeoutSeconds <= 0 {
		connectTimeoutSeconds = 10
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(connectTimeoutSeconds)*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	if username != "" || password != "" {
		clientOpts.SetAuth(options.Credential{
			Username:   username,
			Password:   password,
			AuthSource: authSource,
		})
	}
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}
	// Ping to verify connectivity
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	mongoInstance = &Mongo{
		Client:   client,
		Database: client.Database(database),
	}
	return mongoInstance, nil
}

// GetMongo returns the initialized Mongo singleton, or nil if not initialized.
func GetMongo() *Mongo {
	return mongoInstance
}

// Close terminates the underlying Mongo client.
func (m *Mongo) Close(ctx context.Context) error {
	if m == nil || m.Client == nil {
		return nil
	}
	return m.Client.Disconnect(ctx)
}
