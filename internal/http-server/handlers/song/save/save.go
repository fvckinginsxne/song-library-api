package save

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/render"
	"github.com/go-playground/validator"

	resp "song-library/internal/lib/api/response"
	"song-library/internal/lib/logger/sl"
	"song-library/internal/service/genius"
)

type Request struct {
	Artist string `json:"artist" validate:"required"`
	Title  string `json:"title" validate:"required"`
}

type Response struct {
	resp.Response
	ReleaseDate string `json:"release_date"`
	Lyrics      string `json:"lyrics"`
}

func New(ctx context.Context, log *slog.Logger, client *genius.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.song.save.New"

		logger := log.With(
			slog.String("op", op),
		)

		logger.Info("saving song")

		var req Request

		if err := render.DecodeJSON(r.Body, &req); err != nil {
			logger.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		logger.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			logger.Error("invalid request", sl.Err(err))

			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		trackInfo, err := client.TrackInfo(ctx, req.Artist, req.Title)
		if err != nil {
			if errors.Is(err, genius.ErrTrackNotFound) {
				logger.Error("track not found", sl.Err(err))

				render.JSON(w, r, resp.Error("track not found"))

				return
			}

			logger.Error("failed to fetch track info", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to fetch track info"))

			return
		}

		logger.Info("song saved successfully")

		render.JSON(w, r, Response{
			Response:    resp.OK(),
			ReleaseDate: trackInfo.ReleaseDate,
			Lyrics:      trackInfo.Lyrics,
		})
	}
}
