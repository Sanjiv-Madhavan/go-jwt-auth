package cache

import (
	"log/slog"

	redis "github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
	logger *slog.Logger
}

func NewRedisClient(logger *slog.Logger) *RedisClient {
	client := redis.NewClient(
		&redis.Options{
			Addr:     "redis-19587.c12.us-east-1-4.ec2.redns.redis-cloud.com:19587",
			DB:       0,
			Password: "2GZ3ub9L0YfED448XEB6JkC6Yi1JdIBE",
		},
	)
	return &RedisClient{
		Client: client,
		logger: logger,
	}
}
