package request

//go:generate easyjson $GOFILE

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/mailru/easyjson"
)

// ошибки для данного файла
var (
	ErrBodyTooLarge   = errors.New("request body is too large")
	ErrInvalidJSON    = errors.New("request body contains badly-formed JSON")
	ErrEmptyBody      = errors.New("request body is empty")
	ErrNotOnlyJSONVal = errors.New("body must only contain a single JSON value")
)

// функция декодирования JSON запроса в переданный объект v
func JSON(r *http.Request, v any) error {
	// ограничиваем размер тела запроса - 1 Мб
	const maxBodySize = 1024 * 1024
	// обертка над телом запроса, ограничивает количество байт, которые можно прочитать
	r.Body = http.MaxBytesReader(nil, r.Body, maxBodySize)

	if m, ok := v.(easyjson.Unmarshaler); ok {
		if err := easyjson.UnmarshalFromReader(r.Body, m); err != nil {
			if errors.Is(err, io.EOF) {
				return ErrEmptyBody
			}

			var maxBytesError *http.MaxBytesError
			if errors.As(err, &maxBytesError) {
				return ErrBodyTooLarge
			}
			return fmt.Errorf("%w: %v", ErrInvalidJSON, err)
		}

		return nil
	}

	log.Printf("[WARN] request.JSON fallback! Используем json.Decoder для типа: %T", v)

	// создает объект-декодер, который читает данные напрямую из потока (r.Body) по частям
	decoder := json.NewDecoder(r.Body)

	// читает поток JSON, сопоставляет ключи
	// с тегами структур DTO и выполняет преобразование типов
	if err := decoder.Decode(v); err != nil {
		// Подготавливаем переменные для динамических типов ошибок
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var maxBytesError *http.MaxBytesError

		switch {
		// json.SyntaxError возникает, если JSON поврежден
		case errors.As(err, &syntaxError):
			return fmt.Errorf("%w: at character %d", ErrInvalidJSON, syntaxError.Offset)

		// данные закончились раньше, чем закрылся JSON-объект
		case errors.Is(err, io.ErrUnexpectedEOF):
			return ErrInvalidJSON

		// тип данных в JSON не совпадает с полем структуры
		case errors.As(err, &unmarshalTypeError):
			// проверяем поле .Field, чтобы сообщить фронтенду имя сломанного поля
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return err

		// если Decode возвращает EOF сразу, значит тело было пустым
		case errors.Is(err, io.EOF):
			return ErrEmptyBody

		// размер прочитанных данных превысил установленный лимит в 1МБ
		case errors.As(err, &maxBytesError):
			return ErrBodyTooLarge

		default:
			return err
		}
	}

	// Пробуем прочитать что-то еще. Если там не EOF - в теле мусор.
	err := decoder.Decode(&struct{}{})
	if err != io.EOF {
		return ErrNotOnlyJSONVal
	}
	return nil
}
