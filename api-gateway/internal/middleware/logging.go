package middleware

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
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
	size                int
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// ВОТ ОН! Магический метод, который позволяет "угнать" соединение для WebSockets
func (rw *responseWriterWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("the ResponseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}

func (m *LoggingMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK} // Создаем обертку

		// передаем управление дальше, подменяя http.ResponseWriter на наш wrapped
		next.ServeHTTP(wrapped, r)
		// бизнес-логика выполнилась, ответ сформирован

		duration := time.Since(start)

		m.logger.WithContext(r.Context()).Info("http request finished",
			logger.Int("status", wrapped.statusCode),
			logger.Int("size_bytes", wrapped.size),
			logger.String("duration", duration.String()),
			logger.String("method", r.Method),
			logger.String("path", r.URL.Path),
			logger.String("user_agent", r.UserAgent()),
			logger.String("ip", r.RemoteAddr),
		)
	})
}
