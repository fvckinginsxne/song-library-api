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
	const op = "storage.postgres.Save"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	var songID int64
	row := tx.QueryRowContext(ctx, `
				INSERT INTO songs (artist, title) 
				VALUES ($1, $2) RETURNING id
	`, trackInfo.Artist, trackInfo.Title)

	if err := row.Scan(&songID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = tx.ExecContext(ctx, `
				INSERT INTO song_details (song_id, release_date, lyrics)
				VALUES ($1, $2, $3)
	`, songID, trackInfo.ReleaseDate, trackInfo.Lyrics)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return tx.Commit()
}

func (s *Storage) Close(ctx context.Context) error {
	done := make(chan struct{})
	var closeErr error
	go func() {
		closeErr = s.db.Close()
		close(done)
	}()

	select {
	case <-done:
		return closeErr
	case <-ctx.Done():
		return ctx.Err()
	}
}
