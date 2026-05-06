package logger

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger_Fields(t *testing.T) {
	errTest := errors.New("test error")

	tests := []struct {
		name     string
		fn       func() Field
		expected Field
	}{
		{
			name:     "Успешное создание String",
			fn:       func() Field { return String("str_key", "value") },
			expected: Field{Key: "str_key", Value: "value"},
		},
		{
			name:     "Успешное создание Int",
			fn:       func() Field { return Int("int_key", 42) },
			expected: Field{Key: "int_key", Value: 42},
		},
		{
			name:     "Успешное создание Int64",
			fn:       func() Field { return Int64("int64_key", 100500) },
			expected: Field{Key: "int64_key", Value: int64(100500)},
		},
		{
			name:     "Успешное создание Any",
			fn:       func() Field { return Any("any_key", []int{1, 2, 3}) },
			expected: Field{Key: "any_key", Value: []int{1, 2, 3}},
		},
		{
			name:     "Успешное создание Err",
			fn:       func() Field { return Err(errTest) },
			expected: Field{Key: "error", Value: errTest},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()
			assert.Equal(t, tt.expected, result)
		})
	}
}
