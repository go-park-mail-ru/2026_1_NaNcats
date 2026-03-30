package logger

import (
	"fmt"

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

// Универсальный метод добавления полей
func (l *ZapLogger) With(fields ...zap.Field) *ZapLogger {
	return &ZapLogger{
		logger: l.logger.With(fields...),
	}
}

func (l *ZapLogger) Info(msg string, fields map[string]any) {
	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}

	l.logger.Info(msg, zapFields...)
}

func (l *ZapLogger) Warn(msg string, fields map[string]any) {
	zapFields := make([]zap.Field, 0, len(fields))
	for key, val := range fields {
		zapFields = append(zapFields, zap.Any(key, val))
	}

	l.logger.Warn(msg, zapFields...)
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
