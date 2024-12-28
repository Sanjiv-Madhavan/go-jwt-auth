package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/sanjiv-madhavan/go-jwt-auth/cache"
)

type Middleware struct {
	logger      slog.Logger
	redisClient *cache.RedisClient
}

func NewMiddleware(client *cache.RedisClient, logger *slog.Logger) *Middleware {
	return &Middleware{
		redisClient: client,
		logger:      slog.Logger{},
	}
}

func (m *Middleware) SendJSONResponse(w http.ResponseWriter, statusCode int, v interface{}) {
	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(true)
	if err := encoder.Encode(v); err != nil {
		m.logger.Error("Failed to enode Buffer bytes to JSON", slog.Any("err: ", err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Strict-Transport-Security", "max-age=31353600, includeSubDomains")
	w.Header().Set("Content-Security-Policy", "default-src 'self'")
	w.WriteHeader(statusCode)
	w.Write(buf.Bytes())
}

func (m *Middleware) PanicRecoveryHandler(inner http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				m.logger.Error("Panic recovery failed", slog.Any("error: ", err))
				http.Error(w, "Error occured during panic recovery phase", http.StatusInternalServerError)
			}
		}()
		inner.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
