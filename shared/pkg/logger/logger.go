package logger

import "context"

//go:generate mockgen -source=logger.go -destination=mocks/logger.go -package=mocks Logger

type Field struct {
	Key   string
	Value any
}

func String(key, val string) Field      { return Field{Key: key, Value: val} }
func Int(key string, val int) Field     { return Field{Key: key, Value: val} }
func Int64(key string, val int64) Field { return Field{Key: key, Value: val} }
func Any(key string, val any) Field     { return Field{Key: key, Value: val} }
func Err(err error) Field               { return Field{Key: "error", Value: err} }

type Logger interface {
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Debug(msg string, fields ...Field)
	Error(msg string, err error, fields ...Field)
	Fatal(msg string, err error)
	WithContext(ctx context.Context) Logger
}
