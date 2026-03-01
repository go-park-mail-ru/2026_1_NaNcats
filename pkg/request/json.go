package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func JSON(r *http.Request, v any) error {
	const maxBodySize = 1024 * 1024
	r.Body = http.MaxBytesReader(nil, r.Body, maxBodySize)

	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(v); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("%w: at character %d", ErrInvalidJSON, syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return ErrInvalidJSON

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return err

		case errors.Is(err, io.EOF):
			return ErrEmptyBody

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
