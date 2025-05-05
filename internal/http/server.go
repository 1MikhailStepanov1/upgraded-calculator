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
	router.Post("/execute", func(w http.ResponseWriter, r *http.Request) {
		// TODO убрать перевод в байты вместе с шаблоном фасада
		bodyInBytes, err := io.ReadAll(r.Body)
		response, err := calculator.ExecuteHTTP(ctx, bodyInBytes)
		if err != nil {
			// TODO Сделать отсылку разных кодов ответа на разные ошибки
			w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(response)
	})

	router.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Access-control-allow-origin", "*")
		_, err := w.Write([]byte(swaggerUIHTML()))
		if err != nil {
			logger.Error("Failed to write docs page:", err.Error())
		}
	})

	// Creating server instance
	server := &http.Server{Addr: fmt.Sprintf("0.0.0.0:%d", config.App.HTTPPort), Handler: router}

	// Graceful shutdown handler
	serverCtx, serverStop := context.WithCancel(ctx)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig
		shutdownCtx, _ := context.WithTimeout(serverCtx, config.App.HTTPShutdownTimeout*time.Second)
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

func swaggerUIHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Calculator API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css">
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js"></script>
    <script>
        window.onload = function() {
            window.ui = SwaggerUIBundle({
                url: "/swagger.json",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIBundle.SwaggerUIStandalonePreset
                ],
                layout: "BaseLayout",

            });
        };
    </script>
</body>
</html>`
}
