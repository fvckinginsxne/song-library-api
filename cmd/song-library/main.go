package main

import (
	"fmt"
	"log/slog"
	"os"

	"song-library/internal/config"
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

	log.Info("starting service")

	dbURL := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		cfg.DB.Username, cfg.DB.Password, cfg.DB.Host, cfg.DB.Name)

	storage, err := postgres.New(dbURL)
	if err != nil {
		panic(err.Error())
	}

	client := genius.New(log,
		cfg.GeniusAPI.BaseURL,
		cfg.GeniusAPI.AccessToken,
		storage,
	)

	_ = client

	// TODO: init router: chi, render

	// TODO: run server

	log.Info("service stopped")
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
