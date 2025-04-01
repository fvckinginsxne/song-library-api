package delete

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	resp "song-library/internal/lib/api/response"
	"song-library/internal/lib/logger/sl"
	"song-library/internal/storage"
)

type TrackDeleter interface {
	DeleteTrack(ctx context.Context, uuid string) error
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

		if err := trackDeleter.DeleteTrack(r.Context(), uuid); err != nil {
			if errors.Is(err, storage.ErrInvalidUUID) {
				log.Error("invalid uuid")

				w.WriteHeader(http.StatusBadRequest)

				render.JSON(w, r, resp.Error("invalid request"))
				return
			}

			log.Error("failed to delete track", sl.Err(err))

			w.WriteHeader(http.StatusInternalServerError)

			render.JSON(w, r, resp.Error("internal server error"))
			return
		}

		log.Info("track deleted successfully")

		w.WriteHeader(http.StatusNoContent)
	}
}
