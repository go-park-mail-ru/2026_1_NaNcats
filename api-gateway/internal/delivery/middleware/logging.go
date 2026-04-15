package middleware

import (
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/logger"
)

// Мидлваря для access логов
type LoggingMiddleware struct {
	logger logger.Logger
}

func NewLoggingMiddleware(logger logger.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// Обертка над ResponseWriter, в которой переопределили метод WriteHeader, чтобы мы могли
// запоминать statusCode (чего оригинальный WriteHeader не делает)
type responseWriterWrapper struct {
	http.ResponseWriter // встраивание, чтобы соответствовать интерфейсу ResponseWriter
	statusCode          int
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (m *LoggingMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseWriterWrapper{w, http.StatusOK} // Создаем обертку

		// передаем управление дальше, подменяя http.ResponseWriter на наш wrapped
		next.ServeHTTP(wrapped, r)
		// бизнес-логика выполнилась, ответ сформирован

		duration := time.Since(start)

		userAgent := r.UserAgent()

		m.logger.WithContext(r.Context()).Info("request finished",
			logger.Int("status", wrapped.statusCode),
			logger.String("duration", duration.String()),
			logger.String("method", r.Method),
			logger.String("path", r.URL.Path),
			logger.String("user_agent", userAgent),
			logger.String("ip", r.RemoteAddr),
		)
	})
}
