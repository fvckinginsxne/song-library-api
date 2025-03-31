package save

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/render"
	"github.com/go-playground/validator"

	"song-library/internal/domain/models"
	resp "song-library/internal/lib/api/response"
	"song-library/internal/lib/logger/sl"
	"song-library/internal/service/genius"
)

type Request struct {
	Artist string `json:"artist" validate:"required"`
	Title  string `json:"title" validate:"required"`
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.3 --name=TrackInfoFetcher
type TrackInfoFetcher interface {
	TrackInfo(ctx context.Context, artist, title string) (*models.TrackInfo, error)
}

func New(ctx context.Context, log *slog.Logger, trackInfoFetcher TrackInfoFetcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.song.save.New"

		logger := log.With(
			slog.String("op", op),
		)

		logger.Info("saving song")

		var req Request

		if err := render.DecodeJSON(r.Body, &req); err != nil {
			logger.Error("failed to decode request body", sl.Err(err))

			w.WriteHeader(http.StatusBadRequest)

			render.JSON(w, r, resp.Error("invalid request"))
			return
		}

		logger.Debug("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			logger.Error("invalid request", sl.Err(err))

			w.WriteHeader(http.StatusBadRequest)

			render.JSON(w, r, resp.Error("invalid request"))
			return
		}

		trackInfo, err := trackInfoFetcher.TrackInfo(ctx, req.Artist, req.Title)
		if err != nil {
			if errors.Is(err, genius.ErrTrackNotFound) {
				logger.Error("track not found", sl.Err(err))

				w.WriteHeader(http.StatusNotFound)

				render.JSON(w, r, resp.Error("track not found"))
				return
			}

			logger.Error("failed to fetch track info", sl.Err(err))

			w.WriteHeader(http.StatusInternalServerError)

			render.JSON(w, r, resp.Error("internal error"))
			return
		}

		logger.Info("song saved successfully")

		w.WriteHeader(http.StatusCreated)

		render.JSON(w, r, trackInfo)
	}
}
