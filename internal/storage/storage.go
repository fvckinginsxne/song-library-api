package storage

import "errors"

var (
	ErrTrackNotFound        = errors.New("track not found")
	ErrArtistTracksNotFound = errors.New("artist's tracks not found")
	ErrInvalidUUID          = errors.New("invalid uuid")
	ErrTrackNotCahed        = errors.New("track not cahed")
	ErrArtistTracksNotCahed = errors.New("artist's track not cached")
)
