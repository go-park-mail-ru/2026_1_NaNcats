package validatorutil

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

func FormatValidationError(err error) string {
	if errs, ok := err.(validator.ValidationErrors); ok {
		var errMsgs []string
		for _, e := range errs {
			switch e.Tag() {
			case "min":
				errMsgs = append(errMsgs, e.Field()+" is too short")
			case "max":
				errMsgs = append(errMsgs, e.Field()+" is too long")
			}
		}
		return strings.Join(errMsgs, ",")
	}

	return err.Error()
}
