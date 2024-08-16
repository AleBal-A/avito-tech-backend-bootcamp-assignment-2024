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
	fmt.Println("Starting of estate-service")

	// load config
	cfg := config.MustLoad()
	fmt.Println(cfg)

	// setup logger
	log := logger.SetupLogger(cfg.Logger.Level)
	log.Info("Avito Real-Estate service started", slog.String("env", cfg.Logger.Level))
	log.Debug("Debug message for test")
	log.Error("error message are enabled")

	// DB connection
	conn, err := storage.New(cfg)
	if err != nil {
		log.Error("Could not connect to the database", "error", err)
		panic(err)
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			panic(err)
		}
	}()

	authH, houseH, flatH := setup.InitLayers(conn, cfg, log)
	router := setup.SetupRouter(authH, houseH, flatH, log)

	log.Info("Starting server on port ", slog.String("port", cfg.Server.Port))

	err = http.ListenAndServe(":"+cfg.Server.Port, router)
	if err != nil {
		log.Error("Server startup error", "error", err)
	}
}

/*
	export JWT_SECRET="my_secret_key"
	export DB_USER="user"
	export DB_PASSWORD="password"
*/
