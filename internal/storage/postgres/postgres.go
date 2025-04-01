package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"

	"song-library/internal/domain/models"
	"song-library/internal/storage"
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

func (s *Storage) SaveTrack(ctx context.Context, trackInfo *models.TrackInfo) error {
	const op = "storage.postgres.Save"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO songs (artist, title, release_date, lyrics)
		VALUES ($1, $2, $3, $4)
	`, trackInfo.Artist, trackInfo.Title, trackInfo.ReleaseDate, trackInfo.Lyrics)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return tx.Commit()
}

func (s *Storage) TrackInfo(ctx context.Context, artist, title string) (*models.TrackInfo, error) {
	const op = "storage.postgres.TrackInfo"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	row := tx.QueryRowContext(ctx, `
		SELECT artist, title, release_date, lyrics FROM songs 
		WHERE artist ILIKE $1 AND title ILIKE $2
	`, artist, title)

	var releaseDate, lyrics string

	err = row.Scan(&artist, &title, &releaseDate, &lyrics)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrTrackNotFound
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &models.TrackInfo{
		Artist:      artist,
		Title:       title,
		ReleaseDate: releaseDate,
		Lyrics:      lyrics,
	}, nil
}

func (s *Storage) TracksByArtist(ctx context.Context, artist string) ([]*models.TrackInfo, error) {
	const op = "storage.postgres.TracksByArtist"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
		SELECT artist, title, release_date, lyrics 
		FROM songs WHERE artist ILIKE $1
		ORDER BY release_date DESC 
	`, artist)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var tracks []*models.TrackInfo

	for rows.Next() {
		var title, releaseDate, lyrics string

		if err := rows.Scan(&artist, &title, &releaseDate, &lyrics); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		tracks = append(tracks, &models.TrackInfo{
			Artist:      artist,
			Title:       title,
			ReleaseDate: releaseDate,
			Lyrics:      lyrics,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(tracks) == 0 {
		return nil, fmt.Errorf("%s: %w", op, storage.ErrArtistTracksNotFound)
	}

	return tracks, nil
}

func (s *Storage) DeleteTrack(ctx context.Context, uuid string) error {
	const op = "storage.postgres.DeleteTrack"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `DELETE FROM songs WHERE uuid = $1`, uuid)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrInvalidUUID)
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
