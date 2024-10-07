package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sanjiv-madhavan/go-jwt-auth/env"
)

type Server struct {
	server *http.Server
	router *mux.Router
	logger *slog.Logger
}

func NewServer(router *mux.Router, logger *slog.Logger) *Server {
	return &Server{
		server: &http.Server{},
		router: router,
		logger: logger,
	}
}

func (srv *Server) Start(ctx context.Context) error {
	srv.server.Handler = srv.router
	srv.server.Addr = ":" + env.Environment.HTTPServerPort // To Delete
	srv.logger.Info(fmt.Sprintf("Attempting to start server on the port %s", srv.server.Addr))
	go func() error {
		if err := srv.server.ListenAndServe(); err != nil {
			srv.logger.Error("Failed to start the server", err)
			return err
		}
		return nil
	}()
	return nil
}

func (srv *Server) Wait(ctx context.Context) {
	<-ctx.Done()
	srv.Stop()
}

func (srv *Server) Stop() error {
	srv.logger.Info("Shutting down server")
	graceTimeOut := time.Duration(5 * time.Second)
	srv.logger.Info(fmt.Sprintf("Waiting %s seconds for incoming requests to cease", graceTimeOut))
	time.Sleep(graceTimeOut)
	ctxDeadline := time.Duration(10 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), ctxDeadline)
	defer cancel()
	if err := srv.server.Shutdown(ctx); err != nil {
		srv.logger.Error("Failed to shut down the server", err)
		return err
	}
	srv.logger.Info("Server closed gracefully")
	return nil
}
