package http

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	httpSwagger "github.com/swaggo/http-swagger"
	"io"
	"log/slog"
	"net/http"
	"upgraded-calculator/internal/config"
)

func CreateServer(
	config *config.Config,
	logger *slog.Logger,
	ctx context.Context,
) *http.Server {

	calculator := CalculatorHTTP{logger: logger}

	// Initializing router
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Post("/execute", func(w http.ResponseWriter, r *http.Request) {
		bodyInBytes, err := io.ReadAll(r.Body)

		ctx = context.WithValue(ctx, "request_id", uuid.New().String())

		response, err := calculator.Execute(ctx, bodyInBytes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(response)
	})

	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger.json"),
	))

	router.Get("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, config.App.SwaggerPath)
	})

	// Creating server instance
	server := &http.Server{Addr: fmt.Sprintf("0.0.0.0:%d", config.App.HTTPPort), Handler: router}

	return server
}
