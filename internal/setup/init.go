package setup

import (
	"avito/internal/repositories/authRepo"
	"avito/internal/repositories/flatRepo"
	"avito/internal/repositories/houseRepo"
	"database/sql"
	"log/slog"
)

func InitLayerRepo(conn *sql.DB, log *slog.Logger) {
	authRepo := authRepo.NewRepository(conn, log)
	houseRepo := houseRepo.NewRepository(conn, log)
	flatRepo := flatRepo.NewRepository(conn, log)
}
