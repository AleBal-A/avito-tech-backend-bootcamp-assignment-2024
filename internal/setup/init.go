package setup

import (
	"avito/internal/config"
	"avito/internal/handlers/authHandler"
	"avito/internal/handlers/flatHandler"
	"avito/internal/handlers/houseHandler"
	"avito/internal/repositories/authRepo"
	"avito/internal/repositories/flatRepo"
	"avito/internal/repositories/houseRepo"
	"avito/internal/services/authService"
	"avito/internal/services/flatService"
	"avito/internal/services/houseService"
	"database/sql"
	"log/slog"
)

func InitLayers(
	conn *sql.DB,
	cfg *config.Config,
	log *slog.Logger,
) (
	authHandler.AuthHandler,
	houseHandler.HouseHandler,
	flatHandler.FlatHandler,
) {
	authR := authRepo.NewRepository(conn, log)
	houseR := houseRepo.NewRepository(conn, log)
	flatR := flatRepo.NewRepository(conn, log)

	authS := authService.NewService(authR, cfg.Auth.JWTSecret, log)
	houseS := houseService.NewService(houseR, log)
	flatS := flatService.NewService(flatR, log)

	authH := authHandler.NewHandler(authS, log)
	houseH := houseHandler.NewHandler(houseS, log)
	flatH := flatHandler.NewHandler(flatS, log)

	return authH, houseH, flatH
}
