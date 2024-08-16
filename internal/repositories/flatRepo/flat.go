package flatRepo

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"avito/internal/domain/models"
)

type FlatRepo interface {
	CreateFlat(ctx context.Context, flat *models.Flat) (int, error)
	GetFlatsByHouseID(ctx context.Context, houseID int) ([]*models.Flat, error)
	UpdateFlatStatus(ctx context.Context, flatID int, status string, moderatorID *string) (*models.Flat, error)
	//GetFlatByID(ctx context.Context, houseID int, role string) ([]models.Flat, error)
	GetFlatByID(ctx context.Context, flatID int) (*models.Flat, error)
}

type Repository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewRepository(db *sql.DB, logger *slog.Logger) FlatRepo {
	return &Repository{db: db, logger: logger}
}

// CreateFlat - AuthOnly
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

	query := "SELECT id, house_id, flat_number, price, rooms, status FROM flats WHERE house_id = $1"

	rows, err := r.db.QueryContext(ctx, query, houseID)
	if err != nil {
		r.logger.Error("Failed to get flats by house ID", "op", op, "error", err, "houseID", houseID)
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var flats []*models.Flat
	for rows.Next() {
		flat := &models.Flat{}
		if err := rows.Scan(&flat.ID, &flat.HouseID, &flat.FlatNumber, &flat.Price, &flat.Rooms, &flat.Status); err != nil {
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
func (r *Repository) UpdateFlatStatus(ctx context.Context, flatID int, status string, moderatorID *string) (*models.Flat, error) {
	const op = "repository.flat.UpdateFlatStatus"

	query := `
		UPDATE flats
		SET status = $1, moderator_id = $2
		WHERE id = $3
		RETURNING id, house_id, flat_number, price, rooms, status, moderator_id
	`

	var flat models.Flat
	err := r.db.QueryRowContext(ctx, query, status, moderatorID, flatID).Scan(
		&flat.ID,
		&flat.HouseID,
		&flat.FlatNumber,
		&flat.Price,
		&flat.Rooms,
		&flat.Status,
		&flat.ModeratorID,
	)

	if err != nil {
		r.logger.Error("Failed to update and retrieve flat data", slog.String("op", op), "error", err, slog.Int("flatID", flatID))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &flat, nil
}

//func (r *Repository) GetFlatByID(ctx context.Context, houseID int, role string) ([]models.Flat, error) {
//	const op = "repositories.house.GetFlatsByHouseID"
//
//	var query string
//
//	query = `
//        SELECT id, house_id, flat_number, price, rooms, status
//        FROM flats
//        WHERE house_id = $1
//    `
//
//	if role != "moderator" {
//		query += " AND status = 'approved'"
//	}
//
//	rows, err := r.db.QueryContext(ctx, query, houseID)
//	if err != nil {
//		r.logger.Error("Failed to get flats", "op", op, "error", err, "houseID", houseID)
//		return nil, fmt.Errorf("%s: %w", op, err)
//	}
//	defer rows.Close()
//
//	var flats []models.Flat
//
//	for rows.Next() {
//		var flat models.Flat
//		if err := rows.Scan(&flat.ID, &flat.HouseID, &flat.FlatNumber, &flat.Price, &flat.Rooms, &flat.Status); err != nil {
//			r.logger.Error("Failed to scan flat", "op", op, "error", err)
//			return nil, fmt.Errorf("%s: %w", op, err)
//		}
//		flats = append(flats, flat)
//	}
//
//	if err := rows.Err(); err != nil {
//		r.logger.Error("Error during rows iteration", "op", op, "error", err)
//		return nil, fmt.Errorf("%s: %w", op, err)
//	}
//
//	r.logger.Info("Flats retrieved successfully", "op", op, "houseID", houseID, "flats_count", len(flats))
//	return flats, nil
//}

func (r *Repository) GetFlatByID(ctx context.Context, flatID int) (*models.Flat, error) {
	const op = "repository.flat.GetFlatByID"

	query := `
		SELECT id, house_id, flat_number, price, rooms, status, moderator_id
		FROM flats
		WHERE id = $1
	`

	var flat models.Flat
	err := r.db.QueryRowContext(ctx, query, flatID).Scan(
		&flat.ID,
		&flat.HouseID,
		&flat.FlatNumber,
		&flat.Price,
		&flat.Rooms,
		&flat.Status,
		&flat.ModeratorID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn("Flat not found", slog.String("op", op), slog.Int("flatID", flatID))
			return nil, nil
		}
		r.logger.Error("Failed to retrieve flat by ID", slog.String("op", op), "error", err, slog.Int("flatID", flatID))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &flat, nil
}
