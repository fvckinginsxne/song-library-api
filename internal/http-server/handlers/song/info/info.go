package info

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/render"

	"song-library/internal/domain/models"
	resp "song-library/internal/lib/api/response"
	"song-library/internal/lib/logger/sl"
	"song-library/internal/storage"
)

type TrackProvider interface {
	Track(ctx context.Context, artist, title string) (*models.Track, error)
}

type ArtistTracksProvider interface {
	TracksByArtist(ctx context.Context, artist string) ([]*models.Track, error)
}

func New(ctx context.Context,
	log *slog.Logger,
	trackProvider TrackProvider,
	artistTracksProvider ArtistTracksProvider,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.song.info.New"

		log = log.With(slog.String("op", op))

		log.Info("fetching track info")

		query := r.URL.Query()

		artist := query.Get("artist")
		title := query.Get("title")

		if artist == "" {
			log.Error("missing 'artist' parameter")

			w.WriteHeader(http.StatusBadRequest)

			render.JSON(w, r, resp.Error("artist is required"))
			return
		}

		if title == "" {
			tracks, err := artistTracksProvider.TracksByArtist(ctx, artist)
			if err != nil {
				if errors.Is(err, storage.ErrArtistTracksNotFound) {
					log.Error("artist's track not found")

					w.WriteHeader(http.StatusNotFound)

					render.JSON(w, r, resp.Error("artist's tracks not found"))
					return
				}

				log.Error("failed to get tracks by artist", sl.Err(err))

				w.WriteHeader(http.StatusInternalServerError)

				render.JSON(w, r, resp.Error("internal server error"))
				return
			}

			log.Info("artist's tracks got successfully", slog.Any("tracks", tracks))

			w.WriteHeader(http.StatusOK)

			render.JSON(w, r, tracks)
			return
		}

		track, err := trackProvider.Track(ctx, artist, title)
		if err != nil {
			if errors.Is(err, storage.ErrTrackNotFound) {
				log.Error("track not found")

				w.WriteHeader(http.StatusNotFound)

				render.JSON(w, r, resp.Error("track not found"))
				return
			}

			log.Error("failed to get track", sl.Err(err))

			w.WriteHeader(http.StatusInternalServerError)

			render.JSON(w, r, resp.Error("internal server error"))
			return
		}

		log.Info("track info got successfully", slog.Any("trackInfo", track))

		w.WriteHeader(http.StatusOK)

		render.JSON(w, r, track)
	}
}
