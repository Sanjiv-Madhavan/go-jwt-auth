package controllers

import (
	"context"
	"log/slog"

	"github.com/sanjiv-madhavan/go-jwt-auth/middleware"
)

type Controller struct {
	logger     *slog.Logger
	middleware *middleware.Middleware
	ctx        context.Context
}

func NewController(logger *slog.Logger, middleware *middleware.Middleware, context context.Context) *Controller {
	return &Controller{
		logger:     logger,
		middleware: middleware,
		ctx:        context,
	}
}
