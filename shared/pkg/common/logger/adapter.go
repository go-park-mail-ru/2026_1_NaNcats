// internal/infrastructure/logger/adapter.go
package logger

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type LoggerAdapter struct {
	realLogger *logger.ZapLogger
}

func NewLoggerAdapter(zapLog *logger.ZapLogger) logger.Logger {
	return &LoggerAdapter{realLogger: zapLog}
}

func (a *LoggerAdapter) WithContext(ctx context.Context) logger.Logger {
	if ctx == nil {
		return a
	}

	fields := make([]zap.Field, 0, 2)

	// Пытаемся достать request_id если он есть
	if reqID, ok := ctx.Value(common.RequestIDKey).(string); ok && reqID != "" {
		fields = append(fields, zap.String("request_id", reqID))
	}

	// Достаем TraceID из OpenTelemetry
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		fields = append(fields, zap.String("trace_id", span.SpanContext().TraceID().String()))
	}

	// Если ничего не нашли, возвращаем исходный логгер
	if len(fields) == 0 {
		return a
	}

	return &LoggerAdapter{
		realLogger: a.realLogger.With(fields...),
	}
}

// Вспомогательный метод для конвертации типов без лишней рефлексии
func (a *LoggerAdapter) toZap(fields []logger.Field) []zap.Field {
	res := make([]zap.Field, len(fields))
	for i, f := range fields {
		switch v := f.Value.(type) {
		case string:
			res[i] = zap.String(f.Key, v)
		case int:
			res[i] = zap.Int(f.Key, v)
		case int64:
			res[i] = zap.Int64(f.Key, v)
		case error:
			res[i] = zap.NamedError(f.Key, v)
		default:
			res[i] = zap.Any(f.Key, v) // Рефлексия только для структур
		}
	}
	return res
}

func (a *LoggerAdapter) Info(msg string, fields ...logger.Field) {
	a.realLogger.GetInternal().Info(msg, a.toZap(fields)...)
}

func (a *LoggerAdapter) Warn(msg string, fields ...logger.Field) {
	a.realLogger.GetInternal().Warn(msg, a.toZap(fields)...)
}

func (a *LoggerAdapter) Debug(msg string, fields ...logger.Field) {
	a.realLogger.GetInternal().Debug(msg, a.toZap(fields)...)
}

func (a *LoggerAdapter) Error(msg string, err error, fields ...logger.Field) {
	zf := a.toZap(fields)
	zf = append(zf, zap.Error(err))
	a.realLogger.GetInternal().Error(msg, zf...)
}

func (a *LoggerAdapter) Fatal(msg string, err error) {
	a.realLogger.GetInternal().Fatal(msg, zap.Error(err))
}
