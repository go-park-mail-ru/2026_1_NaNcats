package domain

import "context"

type Logger interface {
	Info(msg string, fields map[string]any)
	Warn(msg string, fields map[string]any)
	Error(msg string, err error, fields map[string]any)
	Fatal(msg string, err error)
	WithContext(ctx context.Context) Logger
}
