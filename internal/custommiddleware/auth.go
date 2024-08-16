package custommiddleware

import (
	"avito/internal/domain/models"
	"avito/internal/handlers/authHandler"
	"context"
	"log/slog"
	"net/http"
	"strings"
)

type ContextKey string

const ClaimsContextKey ContextKey = "claims"

func AuthMiddleware(authH authHandler.AuthHandler, logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const op = "middleware.AuthMiddleware"

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.Error("Missing Authorization header", slog.String("op", op))
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				logger.Error("Invalid Authorization header format", slog.String("op", op))
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			claims, err := authH.ValidateToken(tokenString)
			if err != nil {
				logger.Error("Invalid token", slog.String("op", op), "error", err)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			// Put claims in context
			ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

func RoleMiddleware(allowedRoles []string, logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const op = "middleware.RoleMiddleware"

			claims, ok := r.Context().Value(ClaimsContextKey).(*models.Claims)
			if !ok || claims == nil {
				logger.Error("Unauthorized access attempt", slog.String("op", op))
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			roleValid := false
			for _, role := range allowedRoles {
				if strings.EqualFold(role, claims.Role) {
					roleValid = true
					break
				}
			}

			if !roleValid {
				logger.Warn("Forbidden access attempt", slog.String("op", op), slog.String("role", claims.Role))
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
