package logger

import (
	"log/slog"
	"os"
)

const (
	envDebug = "debug"
	envInfo  = "info"
	envProd  = "prod"
)

func SetupLogger(lvl string) *slog.Logger {
	var log *slog.Logger

	switch lvl {
	case envDebug:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envInfo:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}),
		)
	}

	return log
}
