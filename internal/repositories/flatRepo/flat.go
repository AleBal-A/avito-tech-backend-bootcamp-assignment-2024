package flatRepo

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

// CreateFlat - AuthOnly
// TODO: собрать все данные в слое хендлера для возврата инфы о квартире
func (r *Repository) CreateFlat(ctx context.Context, flat *models.Flat) (int, error) {
	const op = "repository.flat.CreateFlat"

	query := `
		INSERT INTO flats (house_id, flat_number, price, rooms, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var flatID int
	err := r.db.QueryRowContext(ctx, query, flat.HouseID, flat.FlatNumber, flat.Price, flat.Rooms, flat.Status).Scan(&flatID)
	if err != nil {
		r.logger.Error("Failed to create flat", "op", op, "error", err, "houseID", flat.HouseID, "flatNumber", flat.FlatNumber)
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return flatID, nil
}

// GetFlatsByHouseID - AuthOnly
func (r *Repository) GetFlatsByHouseID(ctx context.Context, houseID int) ([]*models.Flat, error) {
	const op = "repository.flat.GetFlatsByHouseID"

	query := "SELECT id, house_id, flat_number, price, rooms, status, created_at FROM flats WHERE house_id = $1"

	rows, err := r.db.QueryContext(ctx, query, houseID)
	if err != nil {
		r.logger.Error("Failed to get flats by house ID", "op", op, "error", err, "houseID", houseID)
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var flats []*models.Flat
	for rows.Next() {
		flat := &models.Flat{}
		if err := rows.Scan(&flat.ID, &flat.HouseID, &flat.FlatNumber, &flat.Price, &flat.Rooms, &flat.Status, &flat.CreatedAt); err != nil {
			r.logger.Error("Failed to scan flat", "op", op, "error", err)
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		flats = append(flats, flat)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Rows iteration error", "op", op, "error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return flats, nil
}

// UpdateFlatStatus - OnlyModerator
// TODO: собрать все данные в слое хендлера для возврата инфы о квартире
func (r *Repository) UpdateFlatStatus(ctx context.Context, flatID int, status string) error {
	const op = "repository.flat.UpdateFlatStatus"

	query := "UPDATE flats SET status = $1 WHERE id = $2"
	_, err := r.db.ExecContext(ctx, query, status, flatID)
	if err != nil {
		r.logger.Error("Failed to update flat status", "op", op, "error", err, "flatID", flatID, "status", status)
		return fmt.Errorf("%s: %w", op, err)
	}

	r.logger.Info("Flat status updated success", "op", op, "flatID", flatID, "status", status)
	return nil
}
