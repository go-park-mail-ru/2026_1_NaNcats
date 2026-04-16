package logger

import (
	"context"
)

// Структура-заглушка для тестов
type NopLogger struct{}

func NewNopLogger() Logger { return &NopLogger{} }

func (l *NopLogger) WithContext(ctx context.Context) Logger { return l }

func (l *NopLogger) Info(msg string, fields ...Field) {}

func (l *NopLogger) Warn(msg string, fields ...Field) {}

func (l *NopLogger) Debug(msg string, fields ...Field) {}

func (l *NopLogger) Error(msg string, err error, fields ...Field) {}

func (l *NopLogger) Fatal(msg string, err error) {}
