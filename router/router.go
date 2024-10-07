package router

import (
	"context"
	"log/slog"

	"github.com/gorilla/mux"
	"github.com/sanjiv-madhavan/go-jwt-auth/controllers"
	"github.com/sanjiv-madhavan/go-jwt-auth/middleware"
)

// auth routes dont have auth check middleware whereas user routes do
func CreateMuxRouter(ctx context.Context, logger *slog.Logger) *mux.Router {
	r := mux.NewRouter().SkipClean(true).UseEncodedPath()

	middleware := middleware.NewMiddleware(logger)
	controller := controllers.NewController(logger, middleware, ctx)

	r.HandleFunc("/v1/healthz", controller.HealthCheckHandler).Methods("GET")
	// r.Use(middleware.PanicRecoveryHandler)
	authRoutesSubRouter := r.PathPrefix("/auth").Subrouter()
	authRoutesSubRouter.HandleFunc("/login", controller.Login)
	authRoutesSubRouter.HandleFunc("/signup", controller.Signup)

	userRoutesSubRouter := r.PathPrefix("/users").Subrouter()
	userRoutesSubRouter.Use(middleware.CreateAuthContext)
	userRoutesSubRouter.HandleFunc("/", controller.ListUsers).Methods("GET")
	userRoutesSubRouter.HandleFunc("/password_reset", controller.UpdatePassword).Methods("POST")
	userRoutesSubRouter.HandleFunc("/{user_id}", controller.ListUserById).Methods("GET")
	return r
}
