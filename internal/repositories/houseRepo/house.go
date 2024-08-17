package houseRepo

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"avito/internal/domain/models"
)

type HouseRepo interface {
	CreateHouse(ctx context.Context, house *models.House) error
	SubscribeToHouse(ctx context.Context, houseID int, email string) error
	GetFlatsByHouseID(ctx context.Context, houseID int, role string) ([]models.Flat, error)
}

type Repository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewRepository(db *sql.DB, logger *slog.Logger) HouseRepo {
	return &Repository{db: db, logger: logger}
}

func (r *Repository) CreateHouse(ctx context.Context, house *models.House) error {
	const op = "repositories.house.CreateHouse"

	query := `
		INSERT INTO houses (address, year_built, builder)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, last_flat_added
	`

	err := r.db.QueryRowContext(ctx, query, house.Address, house.YearBuilt, house.Builder).Scan(&house.ID, &house.CreatedAt, &house.LastFlatAdded)
	if err != nil {
		r.logger.Error("Failed to create house", "op", op, "error", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	r.logger.Info("House created", "op", op, "houseID", house.ID, "address", house.Address, "year_built", house.YearBuilt, "created_at", house.CreatedAt)
	return nil
}

func (r *Repository) GetFlatsByHouseID(ctx context.Context, houseID int, role string) ([]models.Flat, error) {
	const op = "repositories.house.GetFlatsByHouseID"

	var query string
	var args []interface{}

	query = `
        SELECT id, house_id, flat_number, price, rooms, status
        FROM flats
        WHERE house_id = $1
    `
	args = append(args, houseID)

	if role != "moderator" {
		query += " AND status = 'approved'"
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to get flats", "op", op, "error", err, "houseID", houseID)
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var flats []models.Flat

	for rows.Next() {
		var flat models.Flat
		if err := rows.Scan(&flat.ID, &flat.HouseID, &flat.FlatNumber, &flat.Price, &flat.Rooms, &flat.Status); err != nil {
			r.logger.Error("Failed to scan flat", "op", op, "error", err)
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		flats = append(flats, flat)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error during rows iteration", "op", op, "error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	r.logger.Info("Flats retrieved successfully", "op", op, "houseID", houseID, "flats_count", len(flats))
	return flats, nil
}

// TODO: /house/{id}/subscribe
func (r *Repository) SubscribeToHouse(ctx context.Context, houseID int, email string) error {
	const op = "repository.house.SubscribeToHouse"
	return nil
}
