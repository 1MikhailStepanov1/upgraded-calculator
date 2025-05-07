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

// @title Upgraded Calculator API
// @version 1.0
// @description API to execute operations of Upgraded calculator service
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.pipipupu.io/support
// @contact.email support@pipipupu.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

func main() {
	// Initializing environment instruments
	config := cfg.New()
	logger := setupLogger(config.App.LogLevel)
	ctx := context.Background()

	//Initializing HTTP server
	httpServer, httpServerContext := calculatorHttpServer.CreateServer(config, logger, ctx)
	logger.Info("HTTP Server started")
	err := httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logger.Error(err.Error())
		logger.Info("Shutting down HTTP server")
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
