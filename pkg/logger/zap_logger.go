package logger

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"go.uber.org/zap"
)

type ZapLogger struct {
	logger *zap.Logger
}

func NewZapLogger() (*ZapLogger, error) {
	cfg := zap.NewProductionConfig()

	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build zap logger: %w", err)
	}
	return &ZapLogger{logger: logger}, nil
}

// сборка всей необходимой метаинформации о запросе из контекста
func (l *ZapLogger) WithContext(ctx context.Context) domain.Logger {
	if ctx == nil {
		return l
	}

	reqID := middleware.GetRequestID(ctx)
	if reqID != "unknown" {
		// Клонируем логгер с добавленным полем
		return &ZapLogger{
			logger: l.logger.With(zap.String("request_id", reqID)),
		}
	}

	return l
}

func (l *ZapLogger) Info(msg string, fields map[string]any) {
	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}

	l.logger.Info(msg, zapFields...)
}

func (l *ZapLogger) Error(msg string, err error, fields map[string]any) {
	zapFields := make([]zap.Field, 0, len(fields)+1)

	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}

	if err != nil {
		zapFields = append(zapFields, zap.Error(err))
	}

	l.logger.Error(msg, zapFields...)
}

func (l *ZapLogger) Fatal(msg string, err error) {
	l.logger.Fatal(msg, zap.Error(err))
}
