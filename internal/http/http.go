package http

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"upgraded-calculator/internal/common"
	"upgraded-calculator/internal/config"
)

func CreateServer(
	config *config.Config,
	logger *slog.Logger,
	ctx context.Context,
) (*http.Server, context.Context) {

	calculator := common.NewCalculatorFacade(logger)

	// Initializing router
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Post("/", func(w http.ResponseWriter, r *http.Request) {
		// TODO убрать перевод в байты вместе с шаблоном фасада
		bodyInBytes, err := io.ReadAll(r.Body)
		response, err := calculator.Execute(ctx, bodyInBytes)
		if err != nil {
			// TODO Сделать отсылку разных кодов ответа на разные ошибки
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(response)
	})

	// Creating server instance
	server := &http.Server{Addr: fmt.Sprintf("0.0.0.0:%d", config.App.HTTPPort), Handler: router}

	// Graceful shutdown handler
	serverCtx, serverStop := context.WithCancel(ctx)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig
		shutdownCtx, _ := context.WithTimeout(serverCtx, 5*time.Second)
		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				logger.Error("Graceful shutdown timed out.. forcing exit.")
			}
		}()

		err := server.Shutdown(shutdownCtx)
		if err != nil {
			logger.Error(err.Error())
		}
		serverStop()
	}()

	return server, serverCtx
}
