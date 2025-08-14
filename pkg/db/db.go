package db

import (
	"context"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DBClient holds mongo db client
type DBClient struct {
	client *mongo.Client
}

// InitDB initializes a connection to db
func InitDB() (*DBClient, error) {
	uri := os.Getenv("MONGO_URI")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	return &DBClient{
		client: client,
	}, nil
}

func (db *DBClient) Ping() error {
	err := db.client.Ping(context.Background(), nil)
	if err != nil {
		return err
	}
	return nil
}
