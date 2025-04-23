package main

import (
	"context"
	"github.com/joho/godotenv"
	"log/slog"
	"os"
	cfg "upgraded-calculator/internal/config"
)

const (
	localLogsLevel      = "LOCAL"
	productionLogsLevel = "PROD"
)

func init() {
	if err := godotenv.Load(); err != nil {
		slog.Error("No .env file found")
	}
}

func main() {
	config := cfg.New()
	logger := setupLogger(config.App.LogLevel)
	ctx := context.Background()
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case localLogsLevel:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case productionLogsLevel:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
