package mocks

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
)

// Структура-заглушка для тестов
type NopLogger struct{}

func NewNopLogger() domain.Logger { return &NopLogger{} }

func (l *NopLogger) WithContext(ctx context.Context) domain.Logger { return l }

func (l *NopLogger) Info(msg string, fields map[string]any) {}

func (l *NopLogger) Warn(msg string, fields map[string]any) {}

func (l *NopLogger) Error(msg string, err error, fields map[string]any) {}

func (l *NopLogger) Fatal(msg string, err error) {}
