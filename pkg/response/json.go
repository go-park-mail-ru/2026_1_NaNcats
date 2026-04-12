package response

//go:generate easyjson $GOFILE

import (
	"encoding/json"
	"net/http"

	"github.com/mailru/easyjson"
)

// ErrorResponse описывает структуру ответа с ошибкой для Swagger
//
//easyjson:json
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

	var bytes []byte
	var err error

	if m, ok := data.(easyjson.Marshaler); ok {
		bytes, err = easyjson.Marshal(m)
	} else {
		bytes, err = json.Marshal(data)
	}

	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":500,"message":"Failed to serialize response"}`))
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
