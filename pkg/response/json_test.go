package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSON(t *testing.T) {
	t.Run("Успешная отправка данных", func(t *testing.T) {
		rec := httptest.NewRecorder()
		data := map[string]string{"result": "success"}
		statusCode := http.StatusOK

		JSON(rec, statusCode, data)

		assert.Equal(t, statusCode, rec.Code)

		// Проверяем заголовок Content-Type
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		// Проверяем тело ответа
		var body map[string]string
		err := json.NewDecoder(rec.Body).Decode(&body)
		assert.NoError(t, err)
		assert.Equal(t, data, body)
	})

	t.Run("Обработка nil данных", func(t *testing.T) {
		rec := httptest.NewRecorder()

		// Передаем nil
		JSON(rec, http.StatusOK, nil)

		// Проверяем, что вместо null пришел пустой объект {}
		// assert.JSONEq сравнивает две JSON-строки, игнорируя пробелы
		assert.JSONEq(t, "{}", rec.Body.String())
	})
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
