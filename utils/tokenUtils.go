package utils

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sanjiv-madhavan/go-jwt-auth/database"
	"github.com/sanjiv-madhavan/go-jwt-auth/env"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AuthClaims struct {
	Email     string `json:"email,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	UserType  string `json:"user_type,omitempty"`
	UID       string `json:"user_id,omitempty"`
	jwt.RegisteredClaims
}

func GenerateAllTokens(email string, firstName string, lastName string, userType string, uid string) (string, string, error) {
	claims := AuthClaims{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		UserType:  userType,
		UID:       uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	refreshClaims := &AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(env.Environment.SECRET_KEY))
	if err != nil {
		tokenGenerationError := errors.New("failed to generate token")
		return "", "", errors.Join(err, tokenGenerationError)
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(env.Environment.SECRET_KEY))
	if err != nil {
		tokenGenerationError := errors.New("failed to generate token")
		return "", "", errors.Join(err, tokenGenerationError)
	}

	return token, refreshToken, nil
}

func UpdateAllTokens(ctx context.Context, logger *slog.Logger, token string, refreshToken string, UID string) error {
	userCollection := database.OpenCollection(*database.NewMongoClient(ctx, logger), "user")
	var updateObject primitive.D

	updateObject = append(updateObject, bson.E{"token", token})
	updateObject = append(updateObject, bson.E{"refresh_token", refreshToken})

	updatedAt, err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	if err != nil {
		logger.Error("Failed to update token", slog.Any("Error: ", err))
		return err
	}
	updateObject = append(updateObject, bson.E{"updated_at", updatedAt})
	upsert := true
	filter := bson.M{"user_id": UID}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}
	_, err = userCollection.UpdateOne(ctx,
		filter,
		bson.D{
			{"$set", updateObject},
		},
		&opt)

	if err != nil {
		logger.Error("Failed to update token", slog.Any("Error: ", err))
		return err
	}
	return nil
}
