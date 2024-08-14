package custommiddleware

import (
	"avito/internal/domain/models"
	"avito/internal/handlers/authHandler"
	"context"
	"log/slog"
	"net/http"
	"strings"
)

type contextKey string

const userContextKey = contextKey("user")

func AuthMiddleware(authHandler authHandler.AuthHandler, logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const op = "middleware.AuthMiddleware"

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.Error("Missing Authorization header", slog.String("op", op))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				logger.Error("Invalid Authorization header format", slog.String("op", op))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := authHandler.ValidateToken(tokenString)
			if err != nil {
				logger.Error("Invalid token", slog.String("op", op), "error", err)
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RoleMiddleware(allowedRoles []string, logger *slog.Logger) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			const op = "middleware.RoleMiddleware"

			claims, ok := r.Context().Value(userContextKey).(*models.Claims)
			if !ok || claims == nil {
				logger.Error("Unauthorized access attempt", slog.String("op", op))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// list of allowed roles
			roleAllowed := false
			for _, role := range allowedRoles {
				if claims.Role == role {
					roleAllowed = true
					break
				}
			}

			if !roleAllowed {
				logger.Warn("Forbidden access attempt", slog.String("op", op), slog.String("role", claims.Role))
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			logger.Info("Access granted", slog.String("op", op), slog.String("role", claims.Role))
			next.ServeHTTP(w, r)
		}
	}
}
