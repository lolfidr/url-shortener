package redirect

import (
	"errors"
	"log/slog"
	"net/http"
	resp "restapiserv/internal/lib/api/response"
	"restapiserv/internal/lib/logger/sl"
	"restapiserv/internal/storage"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		op := "handlers.url.redirect.New" // Идентификатор операции для логов

		// Добавляем в лог информацию об операции и ID запроса
		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias") // получение параметра
		if alias == "" {
			log.Info("alias is empty")

			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		resURL, err := urlGetter.GetURL(alias)

		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", "alias", alias)
			render.Status(r, http.StatusNotFound) // Добавляем статус 404
			render.JSON(w, r, resp.Error("not found"))
			return
		}

		if err != nil {
			log.Error("failed to get url", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))

			return
		}
		log.Info("got url", slog.String("url", resURL))

		http.Redirect(w, r, resURL, http.StatusFound)

	}
}
