package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/marcelofabianov/web"
	"github.com/marcelofabianov/web/middleware"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg, err := web.LoadConfig()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID())
	r.Use(middleware.RealIP())
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.Logger(logger))
	r.Use(chimiddleware.Compress(5))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		web.Success(w, r, http.StatusOK, map[string]string{
			"service": "enrollment",
			"version": "1.0.0",
			"status":  "running",
		})
	})

	r.Get("/health", web.LivenessHandler)
	r.Get("/health/ready", web.ReadinessHandler())

	logger.Info("starting enrollment service",
		"port", cfg.HTTP.Port,
		"service", "enrollment",
	)

	srv := web.NewServer(cfg, logger, r)
	if err := srv.Start(); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}
