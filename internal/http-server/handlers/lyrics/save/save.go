package save

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/render"
	"github.com/go-playground/validator"

	"lyrics-library/internal/domain/models"
	resp "lyrics-library/internal/lib/api/response"
	"lyrics-library/internal/lib/logger/sl"
	"lyrics-library/internal/service/api"
)

type Request struct {
	Artist string `json:"artist" validate:"required"`
	Title  string `json:"title" validate:"required"`
}

type LyricsFetcher interface {
	Lyrics(ctx context.Context, artist, title string) ([]string, error)
}

type LyricsTranslator interface {
	TranslateLyrics(ctx context.Context, lyrics []string) ([]string, error)
}

type TrackSaver interface {
	SaveTrack(ctx context.Context, track *models.Track) error
}

type TrackCache interface {
	GetTrack(ctx context.Context, artist, title string) (*models.Track, error)
	SaveTrack(ctx context.Context, track *models.Track) error
}

func New(ctx context.Context,
	log *slog.Logger,
	lyricsFetcher LyricsFetcher,
	lyricsTranslator LyricsTranslator,
	trackSaver TrackSaver,
	trackCache TrackCache,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.song.save.New"

		log := log.With(
			slog.String("op", op),
		)

		log.Info("saving lyrics")

		var req Request

		if err := render.DecodeJSON(r.Body, &req); err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			w.WriteHeader(http.StatusBadRequest)

			render.JSON(w, r, resp.Error("invalid request"))
			return
		}

		log.Debug("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			log.Error("invalid request", sl.Err(err))

			w.WriteHeader(http.StatusBadRequest)

			render.JSON(w, r, resp.Error("invalid request"))
			return
		}

		cached, err := trackCache.GetTrack(ctx, req.Artist, req.Title)
		if err == nil {
			log.Info("returnig cached track")

			w.WriteHeader(http.StatusOK)

			render.JSON(w, r, cached)
			return
		}

		lyrics, err := lyricsFetcher.Lyrics(ctx, req.Artist, req.Title)
		if err != nil {
			if errors.Is(err, api.ErrTrackNotFound) {
				log.Error("lyrics not found", sl.Err(err))

				w.WriteHeader(http.StatusNotFound)

				render.JSON(w, r, resp.Error("lyrics not found"))
				return
			}

			log.Error("failed to fetch lyrics", sl.Err(err))

			w.WriteHeader(http.StatusInternalServerError)

			render.JSON(w, r, resp.Error("internal error"))
			return
		}

		log.Debug("lyrics fetched", slog.Any("lyrics", lyrics))

		translation, err := lyricsTranslator.TranslateLyrics(ctx, lyrics)
		if err != nil {

			log.Error("failed translate lyrics", sl.Err(err))

			if errors.Is(err, api.ErrFailedTranslateLyrics) {

				w.WriteHeader(http.StatusBadRequest)

				render.JSON(w, r, resp.Error("failed translate lyrics"))
				return
			}

			w.WriteHeader(http.StatusInternalServerError)

			render.JSON(w, r, resp.Error("internal error"))
			return
		}

		track := &models.Track{
			Artist:      req.Artist,
			Title:       req.Title,
			Lyrics:      lyrics,
			Translation: translation,
		}

		if err := trackSaver.SaveTrack(ctx, track); err != nil {
			log.Error("failed to save track", sl.Err(err))

			w.WriteHeader(http.StatusInternalServerError)

			render.JSON(w, r, resp.Error("internal error"))
			return
		}

		go func() {
			if err := trackCache.SaveTrack(ctx, track); err != nil {
				log.Error("failed to cache track", sl.Err(err))
			}
		}()

		log.Info("lyrics saved successfully")

		w.WriteHeader(http.StatusCreated)

		render.JSON(w, r, track)
	}
}
