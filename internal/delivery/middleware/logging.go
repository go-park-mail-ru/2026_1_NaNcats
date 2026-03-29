package middleware

import (
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
)

// Мидлваря для access логов
type LoggingMiddleware struct {
	logger domain.Logger
}

func NewLoggingMiddleware(logger domain.Logger) *LoggingMiddleware {
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

		m.logger.WithContext(r.Context()).Info("request finished", map[string]any{
			"status":   wrapped.statusCode,
			"duration": duration,
			"method":   r.Method,
			"path":     r.URL.Path,
		})
	})
}
