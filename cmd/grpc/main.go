package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log/slog"
	"net"
	"os"
	cfg "upgraded-calculator/internal/config"
	calculatorGrpcServer "upgraded-calculator/internal/grpc"
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
		slog.Error("No .env file found")
	}
}
func main() {
	config := cfg.New()
	logger := setupLogger(config.App.LogLevel)
	var err error
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", config.App.GRPCPort))
	if err != nil {
		logger.Error("Failed to listen GRPC server: %v", err.Error())
	}

	server := calculatorGrpcServer.CreateServer(config, logger)
	logger.Info("GRPC Server started")
	if err = server.Serve(lis); err != nil {
		logger.Error("Failed to serve GRPC server: %v", err.Error())
	}

}
