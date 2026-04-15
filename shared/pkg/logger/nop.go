package mocks

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/logger"
)

// Структура-заглушка для тестов
type NopLogger struct{}

func NewNopLogger() logger.Logger { return &NopLogger{} }

func (l *NopLogger) WithContext(ctx context.Context) logger.Logger { return l }

func (l *NopLogger) Info(msg string, fields ...logger.Field) {}

func (l *NopLogger) Warn(msg string, fields ...logger.Field) {}

func (l *NopLogger) Debug(msg string, fields ...logger.Field) {}

func (l *NopLogger) Error(msg string, err error, fields ...logger.Field) {}

func (l *NopLogger) Fatal(msg string, err error) {}
