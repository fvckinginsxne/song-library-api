package postgres

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"song-library/internal/domain/models"
)

type Storage struct {
	db *sql.DB
}

func New(dbURL string) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Save(ctx context.Context, trackInfo *models.TrackInfo) error {

	return nil
}
