// Объявление тестового пакета (суффикс _test означает, что это тесты)
package save_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"restapiserv/internal/http-server/handlers/url/save"
	"restapiserv/internal/http-server/handlers/url/save/mocks"
	"restapiserv/internal/lib/logger/handlers/slogdiscard"
)

// TestSaveHandler - основной тест для обработчика сохранения URL
func TestSaveHandler(t *testing.T) {
	// Определяем тестовые кейсы
	cases := []struct {
		name      string // Название кейса
		alias     string // Тестовый alias
		url       string // Тестовый URL
		respError string // Ожидаемая ошибка в ответе
		mockError error  // Ошибка, которую должен вернуть mock
	}{
		{
			name:  "Success",
			alias: "test_alias",
			url:   "https://google.com",
		},
		{
			name:  "Empty alias",
			alias: "",
			url:   "https://google.com",
		},
		{
			name:      "Empty URL",
			url:       "",
			alias:     "some_alias",
			respError: "field URL is a required field",
		},
		{
			name:      "Invalid URL",
			url:       "some invalid URL",
			alias:     "some_alias",
			respError: "field URL is not a valid URL",
		},
		{
			name:      "SaveURL Error",
			alias:     "test_alias",
			url:       "https://google.com",
			respError: "failed to add url",
			mockError: errors.New("unexpected error"),
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Создаем mock для URLSaver
			urlSaverMock := mocks.NewURLSaver(t)

			// Настраиваем mock только для успешных кейсов или кейсов с ошибкой сохранения
			if tc.respError == "" || tc.mockError != nil {
				urlSaverMock.On("SaveURL", tc.url, mock.AnythingOfType("string")).
					Return(int64(1), tc.mockError).
					Once()
			}

			handler := save.New(slogdiscard.NewDiscardLogger(), urlSaverMock)

			// Формируем входные данные в формате JSON
			input := fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, tc.url, tc.alias)

			// Создаем тестовый HTTP запрос
			req, err := http.NewRequest(http.MethodPost, "/save", bytes.NewReader([]byte(input)))
			require.NoError(t, err) // Проверяем что не было ошибки при создании запроса

			// Создаем Recorder для записи ответа
			rr := httptest.NewRecorder()
			// Вызываем обработчик
			handler.ServeHTTP(rr, req)

			// Проверяем что статус код ответа 200 OK
			require.Equal(t, rr.Code, http.StatusOK)

			// Читаем тело ответа
			body := rr.Body.String()

			// Декодируем JSON ответ
			var resp save.Response
			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			// Проверяем что ошибка в ответе соответствует ожидаемой
			require.Equal(t, tc.respError, resp.Error)

			// TODO: add more checks
		})
	}
}
