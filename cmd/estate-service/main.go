package main

import (
	"avito/internal/config"
	"avito/internal/lib/logger"
	"avito/internal/setup"
	"avito/internal/storage"
	"fmt"
	"log/slog"
	"net/http"
)

func main() {
	fmt.Println("Launching settings...")

	// load config
	cfg := config.MustLoad()
	fmt.Println("This config Print needs to be removed in PROD...\n", cfg)

	// setup logger
	log := logger.SetupLogger(cfg.Logger.Level)
	log.Info("Real-Estate service loading...", slog.String("env", cfg.Logger.Level))

	// DB connection
	conn, err := storage.New(cfg)
	if err != nil {
		log.Error("Could not connect to the database", "error", err)
		panic(err)
	}
	log.Info("Successfully connected to the database", slog.String("host", cfg.Database.Host),
		slog.String("db_name", cfg.Database.Name),
	)
	defer func() {
		err = conn.Close()
		if err != nil {
			panic(err)
		}
	}()

	authH, houseH, flatH := setup.InitLayers(conn, cfg, log)
	router := setup.SetupRouter(authH, houseH, flatH, log)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.Timeout,
		WriteTimeout: cfg.Server.Timeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	log.Info("Starting server on port ", slog.String("port", cfg.Server.Port))

	err = srv.ListenAndServe()
	if err != nil {
		log.Error("Server startup error", "error", err)
	}
}

/*
	JWT_SECRET="my_secret_key"
	DB_USER="user"
	DB_PASSWORD="password"
*/
