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

func init() {
	if err := godotenv.Load(); err != nil {
		slog.Debug("No .env file found, using environment values")
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
	config := cfg.New()
	logger := setupLogger(config.App.LogLevel)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 2)

	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", config.App.GRPCPort))
		if err != nil {
			errChan <- fmt.Errorf("failed to listen GRPC server: %v", err)
			return
		}

		server := calculatorGrpcServer.CreateServer(config, logger)
		logger.Info("GRPC Server started")
		if err = server.Serve(lis); err != nil {
			errChan <- fmt.Errorf("failed to serve GRPC server: %v", err)
		}
	}()

	go func() {
		httpServer, httpServerContext := calculatorHttpServer.CreateServer(config, logger, ctx)
		logger.Info("HTTP Server started")

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("HTTP server error: %v", err)
		}
		<-httpServerContext.Done()
	}()

	select {
	case err := <-errChan:
		logger.Error(err.Error())
		cancel()
	case <-ctx.Done():
		logger.Info("Shutting down servers...")
	}
}
