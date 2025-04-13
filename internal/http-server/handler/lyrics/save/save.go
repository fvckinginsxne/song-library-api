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
	trackService "lyrics-library/internal/service/track"
)

type Request struct {
	Artist string `json:"artist" validate:"required"`
	Title  string `json:"title" validate:"required"`
}

type TrackSaver interface {
	Save(ctx context.Context, artist, title string) (*models.Track, error)
}

func New(
	ctx context.Context,
	log *slog.Logger,
	trackSaver TrackSaver,
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

		track, err := trackSaver.Save(ctx, req.Artist, req.Title)
		if err != nil {
			switch {
			case errors.Is(err, trackService.ErrLyricsNotFound):
				w.WriteHeader(http.StatusNotFound)

				render.JSON(w, r, resp.Error("lyrics not found"))
				return
			case errors.Is(err, trackService.ErrFailedTranslateLyrics):
				w.WriteHeader(http.StatusBadRequest)

				render.JSON(w, r, resp.Error("failed translate lyrcis"))
				return
			default:
				w.WriteHeader(http.StatusInternalServerError)

				render.JSON(w, r, resp.Error("internal error"))
				return
			}
		}

		w.WriteHeader(http.StatusCreated)

		render.JSON(w, r, track)
	}
}
