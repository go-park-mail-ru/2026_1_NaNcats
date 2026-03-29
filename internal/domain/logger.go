package domain

import "context"

//go:generate mockgen -destination=mocks/logger_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain Logger
type Logger interface {
	Info(msg string, fields map[string]any)
	Error(msg string, err error, fields map[string]any)
	Fatal(msg string, err error)
	WithContext(ctx context.Context) Logger
}
