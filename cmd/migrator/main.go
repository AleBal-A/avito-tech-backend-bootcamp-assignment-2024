package main

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"avito/internal/config"
)

func main() {
	var dbURL, migrationsPath, migrationsTable string

	// EX: --db-url='postgres://user:pass@localhost:5432/db_name'
	flag.StringVar(&dbURL, "db-url", "", "PostgreSQL connection URL")
	// EX: --migrations-path=./migration
	flag.StringVar(&migrationsPath, "migrations-path", "", "Path to migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "Name of migration table")
	flag.Parse()

	cfg := config.MustLoad()

	// Если dbURL не задан флагом - происходит сборка из cfg
	if dbURL == "" {
		dbURL = buildUrlFromCfg(cfg)
		if dbURL == "" {
			panic("db-url is required")
		}
	}

	if migrationsPath == "" {
		panic("migrations-path is required")
	}
	fmt.Println("migrationsPath", migrationsPath)

	connStr := fmt.Sprintf("%s?sslmode=disable", dbURL)

	m, err := migrate.New(
		"file://"+migrationsPath,
		connStr,
	)
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return
		}
		panic(err)
	}

	fmt.Println("migrations applied")
}

func buildUrlFromCfg(cfg *config.Config) string {
	if cfg.Database.User == "" || cfg.Database.Host == "" || cfg.Database.Port == "" || cfg.Database.Name == "" {
		log.Println("One or more database configuration values are missing")
		return ""
	}

	if cfg.Database.Password == "" {
		return fmt.Sprintf("postgres://%s@%s:%s/%s", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
}
