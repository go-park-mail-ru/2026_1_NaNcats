package logger

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

func TestLoggerAdapter_WithContext(t *testing.T) {
	baseZap, err := logger.NewZapLogger("info")
	require.NoError(t, err)
	adapter := NewLoggerAdapter(baseZap)

	tests := []struct {
		name       string
		setupCtx   func() context.Context
		shouldWrap bool
	}{
		{
			name: "Пустой контекст возвращает исходный логгер",
			setupCtx: func() context.Context {
				return context.Background()
			},
			shouldWrap: false,
		},
		{
			name: "Nil контекст возвращает исходный логгер",
			setupCtx: func() context.Context {
				return nil
			},
			shouldWrap: false,
		},
		{
			name: "Контекст с RequestID обогащает логгер",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), common.RequestIDKey, "req-uuid-123")
			},
			shouldWrap: true,
		},
		{
			name: "Контекст с OpenTelemetry TraceID обогащает логгер",
			setupCtx: func() context.Context {
				traceID, _ := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
				spanID, _ := trace.SpanIDFromHex("00f067aa0ba902b7")

				spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
					TraceID: traceID,
					SpanID:  spanID,
				})
				return trace.ContextWithSpanContext(context.Background(), spanCtx)
			},
			shouldWrap: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			newLogger := adapter.WithContext(ctx)

			assert.NotNil(t, newLogger)
			if tt.shouldWrap {
				assert.NotEqual(t, adapter, newLogger)
			} else {
				assert.Equal(t, adapter, newLogger)
			}
		})
	}
}

func TestLoggerAdapter_LoggingMethods(t *testing.T) {
	baseZap, err := logger.NewZapLogger("info")
	require.NoError(t, err)
	adapter := NewLoggerAdapter(baseZap)

	tests := []struct {
		name   string
		action func()
	}{
		{
			name: "Успешный вызов Info с полями разных типов",
			action: func() {
				adapter.Info("test info",
					logger.String("str", "val"),
					logger.Int("int", 1),
					logger.Int64("int64", 2),
					logger.Err(errors.New("test err")),
					logger.Any("any", []string{"a", "b"}),
				)
			},
		},
		{
			name: "Успешный вызов Error",
			action: func() {
				adapter.Error("test error", errors.New("critical fail"), logger.String("module", "test"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, tt.action)
		})
	}
}
