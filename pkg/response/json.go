package response

import (
	"encoding/json"
	"net/http"
)

// функция кодирования в JSON
func JSON(w http.ResponseWriter, statusCode int, data any) {
	// установка заголовка Content-Type
	// сообщаем клиенту, что в теле ответа JSON, чтобы он мог правильно его распарсить
	w.Header().Set("Content-Type", "application/json")
	// отправляем HTTP-статус
	w.WriteHeader(statusCode)

	// защита от отправки пустого тела
	if data == nil {
		data = map[string]string{}
	}

	// кодирование JSON
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// обертка над JSON для стандартизации сообщений об ошибках
func Error(w http.ResponseWriter, statusCode int, message string) {
	JSON(w, statusCode, map[string]string{"error": message})
}
