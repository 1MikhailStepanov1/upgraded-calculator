package main

import (
	"context"
	"github.com/joho/godotenv"
	"log/slog"
	"net/http"
	"os"
	cfg "upgraded-calculator/internal/config"
	calculatorHttpServer "upgraded-calculator/internal/http"
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
	// Initializing environment instruments
	config := cfg.New()
	logger := setupLogger(config.App.LogLevel)
	ctx := context.Background()

	// Initializing HTTP server
	httpServer, httpServerContext := calculatorHttpServer.CreateServer(config, logger, ctx)
	logger.Info("Server started")
	err := httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logger.Error(err.Error())
	}

	<-httpServerContext.Done()
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
