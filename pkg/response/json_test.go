package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSON(t *testing.T) {
	tests := []struct {
		name           string
		data           any
		expectedStatus int
		checkResponse  func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:           "Успешная отправка данных",
			data:           map[string]string{"result": "success"},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				// Проверяем заголовок Content-Type
				assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

				// Проверяем тело ответа
				var body map[string]string
				err := json.NewDecoder(rec.Body).Decode(&body)
				assert.NoError(t, err)

				expectedData := map[string]string{"result": "success"}
				assert.Equal(t, expectedData, body)
			},
		},
		{
			name:           "Обработка nil данных",
			data:           nil,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				// Проверяем, что вместо null пришел пустой объект {}
				// assert.JSONEq сравнивает две JSON-строки, игнорируя пробелы
				assert.JSONEq(t, "{}", rec.Body.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			statusCode := tt.expectedStatus

			// Вызываем тестируемую функцию
			JSON(rec, statusCode, tt.data)

			assert.Equal(t, statusCode, rec.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, rec)
			}
		})
	}
}

func TestError(t *testing.T) {
	t.Run("Стандартная ошибка", func(t *testing.T) {
		rec := httptest.NewRecorder()
		message := "test error message"
		statusCode := http.StatusBadRequest

		Error(rec, statusCode, message)

		// Проверяем статус
		assert.Equal(t, statusCode, rec.Code)

		// Декодируем и проверяем структуру ErrorResponse
		var resp ErrorResponse
		err := json.NewDecoder(rec.Body).Decode(&resp)

		assert.NoError(t, err)
		assert.Equal(t, statusCode, resp.Code)
		assert.Equal(t, message, resp.Message)
	})
}
