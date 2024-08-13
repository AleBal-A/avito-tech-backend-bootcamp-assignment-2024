package main

import (
	"avito/internal/config"
	"avito/internal/lib/logger"
	"fmt"
	"log/slog"
)

func main() {
	fmt.Println("Starting of estate-service")

	// TODO: load config
	cfg := config.MustLoad()
	fmt.Println(cfg)

	// TODO: setup logger
	log := logger.SetupLogger(cfg.Logger.Level)
	log.Info("Avito Real-Estate service started", slog.String("env", cfg.Logger.Level))
	log.Debug("Debug message for test")
	log.Error("error message are enabled")

	// TODO: init storage

	// TODO: init router
}

/*
	export JWT_SECRET="my_secret_key"
	export DB_USER="user"
	export DB_PASSWORD="password"
*/
