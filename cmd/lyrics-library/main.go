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

	"lyrics-library/internal/client/lyricsovh"
	"lyrics-library/internal/client/yandex"
	"lyrics-library/internal/config"
	del "lyrics-library/internal/http-server/handler/lyrics/delete"
	"lyrics-library/internal/http-server/handler/lyrics/get"
	"lyrics-library/internal/http-server/handler/lyrics/save"
	healthchecker "lyrics-library/internal/http-server/middleware/health-checker"
	"lyrics-library/internal/lib/logger/sl"
	"lyrics-library/internal/lib/logger/slogpretty"
	"lyrics-library/internal/service/track"
	"lyrics-library/internal/storage/postgres"
	"lyrics-library/internal/storage/redis"
)

const (
	envLocal = "local"
	envProd  = "prod"

	shutdownTimeout = 15 * time.Second
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
	)
	defer cancel()

	dbURL := connURL(cfg)

	log.Debug("Connecting to database", slog.String("url", dbURL))

	storage, err := postgres.New(dbURL)
	if err != nil {
		panic(err)
	}

	redisHost := redisHost(cfg)

	log.Debug("Connecting to redis", slog.String("host", redisHost))

	redis, err := redis.New(redisHost, cfg.Redis.Password)
	if err != nil {
		panic(err)
	}

	lyricsClient := lyricsovh.New(log)
	translateClient := yandex.New(log, cfg.YandexTranslatorAPI.Key)

	trackService := track.New(log, 
		lyricsClient, 
		translateClient, 
		storage, redis,
	)

	router := chi.NewRouter()

	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(healthchecker.New(log, storage))

	router.Route("/lyrics", func(r chi.Router) {
		r.Post("/", save.New(ctx, log, trackService))
		r.Get("/", get.New(ctx, log, trackService, trackService))
		r.Delete("/{uuid}", del.New(ctx, log, trackService))
	})

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Info("starting server", slog.String("address", cfg.HTTPServer.Address))

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal recieved")
	case err := <-serverErr:
		log.Error("server error", sl.Err(err))
		cancel()
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("failed to shutdown server", sl.Err(err))
	}

	if err := storage.Close(shutdownCtx); err != nil {
		log.Error("failed to close storage", sl.Err(err))
	}

	if err := redis.Close(shutdownCtx); err != nil {
		log.Error("failed to close redis", sl.Err(err))
	}

	log.Info("service stopped gracefully")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettyLogger()
	case envProd:
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

func connURL(cfg *config.Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DB.Username, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name)
}

func redisHost(cfg *config.Config) string {
	return fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)
}
