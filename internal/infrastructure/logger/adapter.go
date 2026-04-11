package logger

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/logger"
	"go.uber.org/zap"
)

// Адаптер для zap логгера из pkg, который реализует domain.Logger.
// Нужен, чтобы zap_logger в pkg не был никак связан с проектом
type LoggerAdapter struct {
	realLogger *logger.ZapLogger
}

func NewLoggerAdapter(zapLog *logger.ZapLogger) domain.Logger {
	return &LoggerAdapter{realLogger: zapLog}
}

// Сборка всей необходимой метаинформации о запросе из контекста
func (a *LoggerAdapter) WithContext(ctx context.Context) domain.Logger {
	reqID, ok := ctx.Value(middleware.RequestIDKey).(string)
	if !ok || reqID == "" {
		return a
	}

	return &LoggerAdapter{
		realLogger: a.realLogger.With(zap.String("request_id", reqID)),
	}
}

// Просто пробрасываем вызовы в реальный логгер
func (a *LoggerAdapter) Info(msg string, fields map[string]any)  { a.realLogger.Info(msg, fields) }
func (a *LoggerAdapter) Warn(msg string, fields map[string]any)  { a.realLogger.Warn(msg, fields) }
func (a *LoggerAdapter) Debug(msg string, fields map[string]any) { a.realLogger.Debug(msg, fields) }
func (a *LoggerAdapter) Error(msg string, err error, fields map[string]any) {
	a.realLogger.Error(msg, err, fields)
}
func (a *LoggerAdapter) Fatal(msg string, err error) { a.realLogger.Fatal(msg, err) }
