package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"song-library/internal/config"
	"song-library/internal/http-server/handlers/song/save"
	"song-library/internal/lib/logger/sl"
	"song-library/internal/lib/logger/slogpretty"
	"song-library/internal/service/genius"
	"song-library/internal/storage/postgres"
)

const (
	EnvLocal = "local"
	EnvProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbURL := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		cfg.DB.Username, cfg.DB.Password, cfg.DB.Host, cfg.DB.Name)

	storage, err := postgres.New(dbURL)
	if err != nil {
		panic(err)
	}

	client := genius.New(log,
		cfg.GeniusAPI.BaseURL,
		cfg.GeniusAPI.AccessToken,
		storage,
	)

	router := chi.NewRouter()

	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/songs", save.New(ctx, log, client))

	log.Info("starting server", slog.String("address", cfg.HTTPServer.Address))

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("failed connect to server")
		}
	}()

	log.Info("server is running")

	<-stop
	log.Info("shutting down server")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("failed to shutdown server", sl.Err(err))
	}

	log.Info("server stopped gracefully")

	log.Info("shutting down storage")

	if err := storage.Close(shutdownCtx); err != nil {
		log.Error("failed to close storage", sl.Err(err))
	}

	log.Info("storage stopped gracefully")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case EnvLocal:
		log = setupPrettyLogger()
	case EnvProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettyLogger() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
