package delete

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	resp "lyrics-library/internal/lib/api/response"
	"lyrics-library/internal/service/track"
)

type TrackDeleter interface {
	Delete(ctx context.Context, uuid string) error
}

func New(ctx context.Context,
	log *slog.Logger,
	trackDeleter TrackDeleter,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.song.delete.New"

		log = log.With("op", op)

		log.Info("deleting track")

		uuid := chi.URLParam(r, "uuid")

		if uuid == "" {
			log.Error("uuid is required")

			w.WriteHeader(http.StatusBadRequest)

			render.JSON(w, r, resp.Error("invalid request"))
			return
		}

		if err := trackDeleter.Delete(ctx, uuid); err != nil {
			if errors.Is(err, track.ErrInvalidUUID) {
				w.WriteHeader(http.StatusBadRequest)

				render.JSON(w, r, resp.Error("invalid uuid"))
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
