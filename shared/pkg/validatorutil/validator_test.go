package validatorutil

import (
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestFormatValidationError(t *testing.T) {
	validate := validator.New()

	type testStruct struct {
		MinField   string `validate:"min=10"`
		MaxField   string `validate:"max=5"`
		OtherField string `validate:"required"`
	}

	getValidationErr := func(v any) error {
		return validate.Struct(v)
	}

	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "Обычная ошибка (не ValidationErrors)",
			err:      errors.New("internal server error"),
			expected: "internal server error",
		},
		{
			name: "Ошибка валидации с тегом min",
			err: getValidationErr(testStruct{
				MinField:   "short",
				MaxField:   "ok",
				OtherField: "not empty",
			}),
			expected: "MinField is too short",
		},
		{
			name: "Ошибка валидации с тегом max",
			err: getValidationErr(testStruct{
				MinField:   "long enough string",
				MaxField:   "too long",
				OtherField: "not empty",
			}),
			expected: "MaxField is too long",
		},
		{
			name: "Несколько ошибок валидации одновременно (min и max)",
			err: getValidationErr(testStruct{
				MinField:   "short",
				MaxField:   "too long",
				OtherField: "not empty",
			}),
			expected: "MinField is too short,MaxField is too long",
		},
		{
			name: "Ошибка валидации с необрабатываемым тегом (required)",
			err: getValidationErr(testStruct{
				MinField:   "long enough string",
				MaxField:   "ok",
				OtherField: "",
			}),
			expected: "",
		},
		{
			name:     "Передан nil",
			err:      nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			result := FormatValidationError(tt.err)

			assert.Equal(t, tt.expected, result)
		})
	}
}
