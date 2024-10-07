package database

import (
	"context"
	"log/slog"

	"github.com/sanjiv-madhavan/go-jwt-auth/env"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DBClient struct {
	logger      *slog.Logger
	mongoClient *mongo.Client
}

func NewMongoClient(ctx context.Context, logger *slog.Logger) *DBClient {
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(env.Environment.MongoDBURL))
	if err != nil {
		logger.Info("Unable to connect to database")
		return nil
	}
	logger.Info("Connected to database")
	return &DBClient{
		logger:      logger,
		mongoClient: mongoClient,
	}
}
