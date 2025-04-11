package get

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/render"

	"lyrics-library/internal/domain/models"
	resp "lyrics-library/internal/lib/api/response"
	trackService "lyrics-library/internal/service/track"
)

type TrackProvider interface {
	Track(ctx context.Context, artist, title string) (*models.Track, error)
}

type ArtistTracksProvider interface {
	ArtistTracks(ctx context.Context, artist string) ([]*models.Track, error)
}

func New(ctx context.Context,
	log *slog.Logger,
	trackProvider TrackProvider,
	artistTracksProvider ArtistTracksProvider,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.song.read.New"

		log = log.With(slog.String("op", op))

		log.Info("getting lyrics")

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
			tracks, err := artistTracksProvider.ArtistTracks(ctx, artist)
			if err != nil {
				if errors.Is(err, trackService.ErrArtistTracksNotFound) {
					w.WriteHeader(http.StatusBadRequest)

					render.JSON(w, r, resp.Error("artist's tracks not found"))
					return
				}

				w.WriteHeader(http.StatusInternalServerError)

				render.JSON(w, r, resp.Error("internal error"))
				return
			}

			w.WriteHeader(http.StatusOK)

			render.JSON(w, r, tracks)
			return
		}

		track, err := trackProvider.Track(ctx, artist, title)
		if err != nil {
			if errors.Is(err, trackService.ErrTrackNotFound) {
				w.WriteHeader(http.StatusBadRequest)

				render.JSON(w, r, resp.Error("track not found"))
				return
			}

			w.WriteHeader(http.StatusInternalServerError)

			render.JSON(w, r, resp.Error("internal error"))
		}

		w.WriteHeader(http.StatusOK)

		render.JSON(w, r, track)
	}
}
