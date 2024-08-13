package houseRepo

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"avito/internal/domain/models"
)

type Repository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewRepository(db *sql.DB, logger *slog.Logger) *Repository {
	return &Repository{db: db, logger: logger}
}

// TODO: onlymoderator
func (r *Repository) CreateHouse(ctx context.Context, house *models.House) error {
	const op = "repositories.house.CreateHouse"

	query := `
		INSERT INTO houses (address, year_built, builder)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query, house.Address, house.YearBuilt, house.Builder).Scan(&house.ID)
	if err != nil {
		r.logger.Error("Failed to create house", "op", op, "error", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	r.logger.Info("House created", "op", op, "houseID", house.ID, "address", house.Address, "year_built", house.YearBuilt)
	return nil
}

// TODO: /house/{id}/subscribe
