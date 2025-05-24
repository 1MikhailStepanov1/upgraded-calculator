package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
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
	grpcServer := calculatorGrpcServer.CreateServer(config, logger)
	httpServer := calculatorHttpServer.CreateServer(config, logger, ctx)

	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", config.App.GRPCPort))
		if err != nil {
			errChan <- fmt.Errorf("failed to listen GRPC server: %v", err)
			return
		}

		logger.Info("GRPC Server started")
		if err = grpcServer.Serve(lis); err != nil {
			errChan <- fmt.Errorf("failed to serve GRPC server: %v", err)
		}
	}()

	go func() {
		logger.Info("HTTP Server started")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("HTTP server error: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig
		logger.Info("Received shutdown signal. Stopping servers...")

		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, config.App.HTTPShutdownTimeout*time.Second)
		defer shutdownCancel()

		var wg sync.WaitGroup
		wg.Add(2)

		// Graceful shutdown HTTP
		go func() {
			defer wg.Done()
			if err := httpServer.Shutdown(shutdownCtx); err != nil {
				logger.Error("HTTP shutdown error:", err)
			} else {
				logger.Info("HTTP Server stopped gracefully")
			}
		}()

		// Graceful shutdown gRPC
		go func() {
			defer wg.Done()
			grpcGracefulDone := make(chan struct{})
			go func() {
				grpcServer.GracefulStop()
				close(grpcGracefulDone)
			}()

			select {
			case <-grpcGracefulDone:
				logger.Info("GRPC Server stopped gracefully")
			case <-time.After(config.App.GRPCShutdownTimeout * time.Second):
				logger.Warn("GRPC graceful shutdown timed out. Forcing stop.")
				grpcServer.Stop()
			}
		}()

		wg.Wait()
		cancel()
	}()

	select {
	case err := <-errChan:
		logger.Error(err.Error())
		cancel()
	case <-ctx.Done():
		logger.Info("Servers stopped")
	}
}
