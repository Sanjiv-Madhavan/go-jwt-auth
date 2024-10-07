package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/sanjiv-madhavan/go-jwt-auth/database"
	"github.com/sanjiv-madhavan/go-jwt-auth/env"
	"github.com/sanjiv-madhavan/go-jwt-auth/router"
	"github.com/sanjiv-madhavan/go-jwt-auth/server"
)

func main() {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true, // Enable adding filename and line number
	}))

	err := env.LoadEnvironment()
	if err != nil {
		logger.Error("Failed to load environment", err)
	}

	database.ConnectToDB(ctx, logger)

	r := router.CreateMuxRouter(ctx, logger)
	server := server.NewServer(r, logger)
	server.Start(ctx)
	server.Wait(ctx)
}
