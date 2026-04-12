// internal/infrastructure/logger/adapter.go
package logger

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/logger"
	"go.uber.org/zap"
)

type LoggerAdapter struct {
	realLogger *logger.ZapLogger
}

func NewLoggerAdapter(zapLog *logger.ZapLogger) domain.Logger {
	return &LoggerAdapter{realLogger: zapLog}
}

func (a *LoggerAdapter) WithContext(ctx context.Context) domain.Logger {
	reqID, ok := ctx.Value(middleware.RequestIDKey).(string)
	if !ok || reqID == "" {
		return a
	}
	// Если ID есть, создаем один раз привязанный логгер
	return &LoggerAdapter{
		realLogger: a.realLogger.With(zap.String("request_id", reqID)),
	}
}

// Вспомогательный метод для конвертации типов без лишней рефлексии
func (a *LoggerAdapter) toZap(fields []domain.Field) []zap.Field {
	res := make([]zap.Field, len(fields))
	for i, f := range fields {
		switch v := f.Value.(type) {
		case string:
			res[i] = zap.String(f.Key, v)
		case int:
			res[i] = zap.Int(f.Key, v)
		case error:
			res[i] = zap.NamedError(f.Key, v)
		default:
			res[i] = zap.Any(f.Key, v) // Рефлексия только для структур
		}
	}
	return res
}

func (a *LoggerAdapter) Info(msg string, fields ...domain.Field) {
	a.realLogger.GetInternal().Info(msg, a.toZap(fields)...)
}

func (a *LoggerAdapter) Warn(msg string, fields ...domain.Field) {
	a.realLogger.GetInternal().Warn(msg, a.toZap(fields)...)
}

func (a *LoggerAdapter) Debug(msg string, fields ...domain.Field) {
	a.realLogger.GetInternal().Debug(msg, a.toZap(fields)...)
}

func (a *LoggerAdapter) Error(msg string, err error, fields ...domain.Field) {
	zf := a.toZap(fields)
	zf = append(zf, zap.Error(err))
	a.realLogger.GetInternal().Error(msg, zf...)
}

func (a *LoggerAdapter) Fatal(msg string, err error) {
	a.realLogger.GetInternal().Fatal(msg, zap.Error(err))
}
