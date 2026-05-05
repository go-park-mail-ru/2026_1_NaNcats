package logger

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNopLogger(t *testing.T) {
	t.Run("Успешное создание и вызов методов NopLogger", func(t *testing.T) {
		nop := NewNopLogger()
		assert.NotNil(t, nop)

		assert.NotPanics(t, func() {
			ctx := context.Background()
			nopCtx := nop.WithContext(ctx)
			assert.Equal(t, nop, nopCtx)

			nop.Info("info message", String("k", "v"))
			nop.Warn("warn message")
			nop.Debug("debug message")
			nop.Error("error message", errors.New("err"))
			nop.Fatal("fatal message", errors.New("fatal"))
		})
	})
}
