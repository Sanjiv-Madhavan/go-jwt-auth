package middleware

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sanjiv-madhavan/go-jwt-auth/cache"
	"github.com/sanjiv-madhavan/go-jwt-auth/constants"
	"github.com/sanjiv-madhavan/go-jwt-auth/env"
	"github.com/sanjiv-madhavan/go-jwt-auth/utils"
)

func (m *Middleware) ValidateToken(authToken string) (*utils.AuthClaims, error) {
	token, err := jwt.ParseWithClaims(authToken, &utils.AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(env.Environment.SECRET_KEY), nil
	})

	if err != nil {
		m.logger.Error("Authorization Error", slog.Any("Error: ", err))
		return nil, errors.New(http.StatusText(http.StatusUnauthorized))
	}

	if token == nil || !token.Valid {
		return nil, errors.New(http.StatusText(http.StatusUnauthorized))
	}

	if claims, ok := token.Claims.(*utils.AuthClaims); ok && token.Valid {
		return claims, nil
	}

	m.logger.Error("Invalid token claims")
	return nil, errors.New(http.StatusText(http.StatusUnauthorized))
}

func (m *Middleware) CreateAuthContext(inner http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=31353600, includeSubDomains")
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			m.logger.Error("Auth token not found")
			m.SendJSONResponse(w, http.StatusUnauthorized, "Auth token not found")
			return
		}

		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		claims, err := m.ValidateToken(tokenString)
		if err != nil {
			m.logger.Error("Auth token invalid", slog.Any("error: ", err))
			m.SendJSONResponse(w, http.StatusUnauthorized, "Auth token invalid")
			return
		}

		globalInvalidationTimestamp, err := cache.GetGlobalInvalidation(r.Context(), m.redisClient)
		if err != nil {
			m.logger.Error("Failed to retrieve global invalidation", slog.Any("Err:", err))
			m.SendJSONResponse(w, http.StatusUnauthorized, "Failed to retrieve global invalidation")
			return
		}
		if globalInvalidationTimestamp > 0 && claims.IssuedAt.Unix() < globalInvalidationTimestamp {
			m.logger.Error("Auth token Expired", slog.Any("Err:", err))
			m.SendJSONResponse(w, http.StatusUnauthorized, "Auth Token Expired")
			return
		}

		userInvalidation, err := cache.GetUserSpecificInvalidation(r.Context(), m.redisClient, claims.UID)
		if err != nil {
			m.logger.Error(fmt.Sprintf("Faied to retrieve token invalidation for User: %s", claims.UID), slog.Any("Err:", err))
			m.SendJSONResponse(w, http.StatusUnauthorized, fmt.Sprintf("Faied to retrieve token invalidation for User: %s", claims.UID))
			return
		}
		if userInvalidation > 0 && claims.IssuedAt.Unix() < userInvalidation {
			m.logger.Error(fmt.Sprintf("Auth token Expired for user: %s", claims.UID), slog.Any("Err:", err))
			m.SendJSONResponse(w, http.StatusUnauthorized, fmt.Sprintf("Auth token Expired for user: %s", claims.UID))
			return
		}

		ctx := context.WithValue(r.Context(), constants.Email, claims.Email)
		ctx = context.WithValue(ctx, constants.FirstName, claims.FirstName)
		ctx = context.WithValue(ctx, constants.LastName, claims.LastName)
		ctx = context.WithValue(ctx, constants.UID, claims.UID)
		ctx = context.WithValue(ctx, constants.UserType, claims.UserType)
		ctx = context.WithValue(ctx, constants.ExpiresAt, claims.RegisteredClaims.ExpiresAt.Unix())
		r = r.WithContext(ctx)
		inner.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
