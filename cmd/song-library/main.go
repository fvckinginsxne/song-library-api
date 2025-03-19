package main

import (
	"log/slog"
	"os"

	"song-library/internal/config"
)

func main() {
	cfg := config.MustLoad()

	log := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)

	log.Info("starting song-library", slog.Any("config", cfg))

	// TODO: init storage: postgres

	// TODO: init router: chi, render

	// TODO: run server
}
