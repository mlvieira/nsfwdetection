package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mlvieira/nsfwdetection/internal/config"
	"github.com/mlvieira/nsfwdetection/internal/models"
	"github.com/mlvieira/nsfwdetection/internal/utils"
)

type ContextKey string

var jwtSecretKey = config.AppConfig.Security.JWTSecretKey

const UserKey = ContextKey("user")

// JWTAuth validates the JWT token in the Authorization header
func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.WriteJSONError(w, http.StatusUnauthorized, "Missing Authorization header")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.WriteJSONError(w, http.StatusUnauthorized, "Invalid Authorization header format")
			return
		}

		tokenStr := parts[1]

		claims := &models.Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecretKey), nil
		})

		if err != nil || !token.Valid {
			utils.WriteJSONError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), UserKey, claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
