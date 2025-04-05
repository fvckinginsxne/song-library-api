package read

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/render"

	"lyrics-library/internal/domain/models"
	resp "lyrics-library/internal/lib/api/response"
	"lyrics-library/internal/lib/logger/sl"
	"lyrics-library/internal/storage"
)

type TrackProvider interface {
	Track(ctx context.Context, artist, title string) (*models.Track, error)
}

type ArtistTracksProvider interface {
	TracksByArtist(ctx context.Context, artist string) ([]*models.Track, error)
}

type TrackCache interface {
	SaveTrack(ctx context.Context, track *models.Track) error
	GetTrack(ctx context.Context, artist, title string) (*models.Track, error)
}

type ArtistTracksCache interface {
	SaveArtistTracks(ctx context.Context, artist string, tracks []*models.Track) error
	GetArtistTracks(ctx context.Context, artist string) ([]*models.Track, error)
}

func New(ctx context.Context,
	log *slog.Logger,
	trackProvider TrackProvider,
	artistTracksProvider ArtistTracksProvider,
	trackCache TrackCache,
	artistTracksCache ArtistTracksCache,
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
			cachedTracks, err := artistTracksCache.GetArtistTracks(ctx, artist)
			if err == nil {
				log.Info("getting tracks from cache")

				w.WriteHeader(http.StatusOK)

				render.JSON(w, r, cachedTracks)
				return
			}

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

				render.JSON(w, r, resp.Error("internal error"))
				return
			}

			go func() {
				if err := artistTracksCache.SaveArtistTracks(ctx, artist, tracks); err != nil {
					log.Error("failed to cache artist tracks", sl.Err(err))
				}
			}()

			log.Info("artist's tracks got successfully", slog.Any("tracks", tracks))

			w.WriteHeader(http.StatusOK)

			render.JSON(w, r, tracks)
			return
		}

		cachedTrack, err := trackCache.GetTrack(ctx, artist, title)
		if err == nil {
			log.Info("getting track from cache")

			w.WriteHeader(http.StatusOK)

			render.JSON(w, r, cachedTrack)
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

			render.JSON(w, r, resp.Error("internal error"))
			return
		}

		go func() {
			if err := trackCache.SaveTrack(ctx, track); err != nil {
				log.Error("failed to cache track", sl.Err(err))
			}
		}()

		log.Info("lyrics got successfully", slog.Any("track", track))

		w.WriteHeader(http.StatusOK)

		render.JSON(w, r, track)
	}
}
