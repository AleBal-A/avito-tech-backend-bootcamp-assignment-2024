package authRepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"log/slog"

	"avito/internal/domain/models"
	"avito/internal/repositories"
)

type AuthRepo interface {
	CreateUser(ctx context.Context, user *models.User) (string, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
}

type Repository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewRepository(db *sql.DB, logger *slog.Logger) *Repository {
	return &Repository{db: db, logger: logger}
}

// CreateUser - Register
func (r *Repository) CreateUser(ctx context.Context, user *models.User) (string, error) {
	const op = "repositories.auth.CreateUser"

	query := "INSERT INTO users (email, password_hash, role) VALUES ($1, $2, $3) RETURNING id"

	var userID string
	err := r.db.QueryRowContext(ctx, query, user.Email, user.Password, user.Role).Scan(&userID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == repositories.UniqueViolation {
			r.logger.Warn("User already exists", "op", op, "email", user.Email)
			return "", fmt.Errorf("%s: %w", op, repositories.ErrUserExists)
		}
		r.logger.Error("Failed to execute statement", "op", op, "error", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return userID, nil
}

// GetUserByEmail - Login
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	const op = "repositories.auth.GetUserByEmail"

	query := "SELECT id, email, password_hash, role FROM users WHERE email = $1"

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Email, &user.Password, &user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Warn("User not found", "op", op, "email", email)
			return nil, fmt.Errorf("%s: %w", op, repositories.ErrUserNotFound)
		}
		r.logger.Error("Failed to query user by email", "op", op, "error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	r.logger.Info("User found", "op", op, "email", email)
	return user, nil
}
