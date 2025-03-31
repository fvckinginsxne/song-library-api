package storage

import "errors"

var (
	ErrTrackNotFound        = errors.New("track not found")
	ErrArtistTracksNotFound = errors.New("artist's tracks not found")
)
