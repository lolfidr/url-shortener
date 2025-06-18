package save

import (
	"errors"
	"io"
	"net/http"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"

	resp "restapiserv/internal/lib/api/response"
	"restapiserv/internal/lib/logger/sl"
	"restapiserv/internal/lib/random"
	"restapiserv/internal/storage"
)

// Request представляет структуру входящего запроса на сохранение URL
type Request struct {
	URL   string `json:"url" validate:"required,url"` // Обязательное поле, должно быть валидным URL
	Alias string `json:"alias,omitempty" validate:"omitempty,alphanum,min=3,max=20"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

// Константа для длины генерируемого алиаса, если пользователь не указал свой
const aliasLength = 6

// URLSaver интерфейс для сохранения URL
type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

// New создает обработчик HTTP для сохранения URL
func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New" // Идентификатор операции для логов

		// Добавляем в лог информацию об операции и ID запроса
		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		// Пытаемся декодировать тело запроса в структуру Request
		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {

			// Обработка случая с пустым телом запроса
			log.Error("request body is empty")
			render.JSON(w, r, resp.Error("empty request"))
			return
		}
		if err != nil {

			// Ошибка декодирования тела запроса
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		// Валидация полей запроса с использованием validator
		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("invalid request", sl.Err(err))
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

		// Если пользователь не указал алиас, генерируем случайный
		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		// Сохраняем URL в хранилище
		id, err := urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			// Обработка случая, когда URL уже существует
			log.Info("url already exists", slog.String("url", req.URL))
			render.JSON(w, r, resp.Error("url already exists"))
			return
		}
		if err != nil {
			// Ошибка при сохранении URL
			log.Error("failed to add url", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to add url"))
			return
		}

		log.Info("url added", slog.Int64("id", id))

		// Отправляем успешный ответ с алиасом
		responseOK(w, r, alias)
	}
}

// responseOK отправляет успешный HTTP-ответ
func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: resp.OK(), // Базовый успешный ответ
		Alias:    alias,     // Алиас (может быть сгенерированным или пользовательским)
	})
}
