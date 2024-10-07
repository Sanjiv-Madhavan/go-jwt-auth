package database

import (
	"context"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func ConnectToDB(ctx context.Context, logger *slog.Logger) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	dbClient := NewMongoClient(ctx, logger)
	defer dbClient.mongoClient.Disconnect(ctx)
	databases, err := dbClient.mongoClient.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		dbClient.logger.Error("Something went wrong while reading databases", slog.Any("error:", err))
	}
	dbClient.logger.Info("Databases available: ", databases)
	return nil
}

func OpenCollection(dbClient DBClient, collectionName string) *mongo.Collection {
	collection := dbClient.mongoClient.Database("Cluster0").Collection(collectionName)
	return collection
}
