package cache

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sanjiv-madhavan/go-jwt-auth/constants"
)

func HealthCheckHandler(ctx context.Context, logger *slog.Logger) {
	r := NewRedisClient(logger)
	pong, err := r.Client.Ping(ctx).Result()
	if err != nil {
		r.logger.Error("Health Check Failed", slog.Any("Error:", err))
		return
	}
	r.logger.Info(fmt.Sprintf("Ping from Redis: %s", pong))
}

func setGlobalInvalidation(ctx context.Context, r *RedisClient, timestamp int64) {
	ttl := time.Duration(24 * time.Hour)
	if err := r.Client.Set(ctx, constants.GlobalInvalidationKey, timestamp, ttl); err != nil {
		r.logger.Error("Global Invalidation Failed", slog.Any("Error: ", err))
		return
	}
}

func SetUserSpecificInvalidation(ctx context.Context, r *RedisClient, userID string, timestamp int64, tokenValidity int64) {
	ttl := time.Duration(tokenValidity)
	key := constants.UserInvalidation + ":" + userID
	if err := r.Client.Set(ctx, key, timestamp, ttl); err != nil {
		r.logger.Error(fmt.Sprintf("Token invalidation for %s user failed", userID), slog.Any("Error: ", err))
		return
	}
}

func GetGlobalInvalidation(ctx context.Context, r *RedisClient) (int64, error) {
	invalidationTimestamp, err := r.Client.Get(ctx, constants.GlobalInvalidationKey).Result()
	if err != nil {
		if err == redis.Nil {
			r.logger.Info("No invalidation event found", slog.Any("Error: ", err))
			return 0, nil
		}
		r.logger.Error("Failed to retrieve global invalidation", slog.Any("Error: ", err))
		return 0, err
	}
	timestamp, err := strconv.ParseInt(invalidationTimestamp, 10, 64)
	if err != nil {
		r.logger.Error("Failed to retrieve global invalidation", slog.Any("Error: ", err))
		return 0, err
	}
	return timestamp, nil
}

func GetUserSpecificInvalidation(ctx context.Context, r *RedisClient, userID string) (int64, error) {
	key := constants.UserInvalidation + ":" + userID
	invalidationTimestamp, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			r.logger.Info("No invalidation event found", slog.Any("Error: ", err))
			return 0, nil
		}
		r.logger.Error(fmt.Sprintf("Failed to retrieve invalidation for user: %s", userID), slog.Any("Error: ", err))
		return 0, err
	}
	timestamp, err := strconv.ParseInt(invalidationTimestamp, 10, 64)
	if err != nil {
		r.logger.Error(fmt.Sprintf("Failed to retrieve invalidation for user: %s", userID), slog.Any("Error: ", err))
		return 0, err
	}
	return timestamp, nil
}
