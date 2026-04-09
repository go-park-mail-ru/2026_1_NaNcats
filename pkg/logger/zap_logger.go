package logger

import (
	"fmt"

	"go.uber.org/zap"
)

type ZapLogger struct {
	logger *zap.Logger
}

func NewZapLogger(levelStr string) (*ZapLogger, error) {
	var cfg zap.Config

	// Формат и стиль логов
	if levelStr == "debug" {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}

	// Какой уровень логов принимать, те что ниже игнорировать
	level, err := zap.ParseAtomicLevel(levelStr)
	if err != nil {
		level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	cfg.Level = level

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
	zapFields := transferFields(fields, len(fields))

	l.logger.Info(msg, zapFields...)
}

func (l *ZapLogger) Warn(msg string, fields map[string]any) {
	zapFields := transferFields(fields, len(fields))

	l.logger.Warn(msg, zapFields...)
}

func (l *ZapLogger) Debug(msg string, fields map[string]any) {
	zapFields := transferFields(fields, len(fields))

	l.logger.Debug(msg, zapFields...)
}

func (l *ZapLogger) Error(msg string, err error, fields map[string]any) {
	zapFields := transferFields(fields, len(fields)+1)

	if err != nil {
		zapFields = append(zapFields, zap.Error(err))
	}

	l.logger.Error(msg, zapFields...)
}

func (l *ZapLogger) Fatal(msg string, err error) {
	l.logger.Fatal(msg, zap.Error(err))
}

func transferFields(fields map[string]any, capacity int) []zap.Field {
	zapFields := make([]zap.Field, 0, capacity)

	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}

	return zapFields
}
