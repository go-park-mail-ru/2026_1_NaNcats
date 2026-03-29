package response

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse описывает структуру ответа с ошибкой для Swagger
type ErrorResponse struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"Неверный формат запроса"`
}

// функция кодирования в JSON
func JSON(w http.ResponseWriter, statusCode int, data any) {
	// защита от отправки пустого тела
	if data == nil {
		data = map[string]string{}
	}

	w.Header().Set("Content-Type", "application/json")

	bytes, err := json.Marshal(data)
	if err != nil {
		resp := []byte(`{"code":500,"message":"Failed to serialize response"}`)

		w.WriteHeader(http.StatusOK)
		w.Write(resp)
		return
	}

	w.WriteHeader(statusCode)
	w.Write(bytes)
}

// обертка над JSON для стандартизации сообщений об ошибках
func Error(w http.ResponseWriter, statusCode int, message string) {
	resp := ErrorResponse{
		Code:    statusCode,
		Message: message,
	}

	networkStatus := statusCode
	if statusCode >= 500 {
		networkStatus = http.StatusOK
	}

	JSON(w, networkStatus, resp)
}
