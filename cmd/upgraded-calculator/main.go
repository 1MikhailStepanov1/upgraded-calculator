package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"log/slog"
	"net"
	"net/http"
	"os"
	cfg "upgraded-calculator/internal/config"
	calculatorGrpcServer "upgraded-calculator/internal/grpc"
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

	errChan := make(chan error)
	//Initializing HTTP server
	go func() {
		httpServer, httpServerContext := calculatorHttpServer.CreateServer(config, logger, ctx)
		logger.Info("HTTP Server started")
		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Error(err.Error())
			errChan <- err
		}
		<-httpServerContext.Done()
	}()

	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", config.App.GRPCPort))
		if err != nil {
			logger.Error("Failed to listen GRPC server: %v", err.Error())
		}

		server := calculatorGrpcServer.CreateServer(config, logger)
		logger.Info("GRPC Server started")
		if err := server.Serve(lis); err != nil {
			logger.Error("Failed to serve GRPC server: %v", err.Error())
		}
	}()

	select {
	case err := <-errChan:
		logger.Error("Server error:", err)
	case <-ctx.Done():
		logger.Info("Shutting down servers")
	}
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
