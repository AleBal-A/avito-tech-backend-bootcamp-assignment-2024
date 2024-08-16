package storage

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"os"

	"avito/internal/config"
)

// New creates a new connection of the SQL storage.
func New(cfg *config.Config) (*sql.DB, error) {
	const op = "database.postgresql.New"

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.Name,
		)
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return db, nil
}
