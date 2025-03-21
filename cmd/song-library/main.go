package main

import (
	"fmt"
	"log/slog"
	"os"

	"song-library/internal/config"
	"song-library/internal/lib/logger/sl"
	"song-library/internal/storage/postgres"
)

func main() {
	cfg := config.MustLoad()

	log := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)

	log.Info("starting service")

	dbURL := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		cfg.DB.Username, cfg.DB.Password, cfg.DB.Host, cfg.DB.Name)

	storage, err := postgres.New(dbURL)
	if err != nil {
		log.Error("failed to connect to database", sl.Err(err))
	}

	_ = storage

	// TODO: init service layer

	// TODO: init router: chi, render

	// TODO: run server
}
