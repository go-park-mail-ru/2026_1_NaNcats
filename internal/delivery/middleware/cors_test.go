package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCORSMiddleware(t *testing.T) {
	allowedOrigins := []string{"http://localhost:2033", "http://foodcourt.fun/"}
	mw := NewCORSMiddleware(allowedOrigins)

	// Флаг для проверки того, дошел ли запрос до следующего хэндлера
	nextCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	t.Run("Разрешенный Origin", func(t *testing.T) {
		nextCalled = false
		origin := "http://localhost:2033"
		req := httptest.NewRequest(http.MethodPost, "/api/any", nil)
		req.Header.Set("Origin", origin)
		rec := httptest.NewRecorder()

		mw.Handler(nextHandler).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, origin, rec.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
		assert.True(t, nextCalled, "Запрос должен пройти в следующий хэндлер")
	})

	t.Run("Запрещенный Origin", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest(http.MethodPost, "/api/any", nil)
		req.Header.Set("Origin", "http://hacker.com")
		rec := httptest.NewRecorder()

		mw.Handler(nextHandler).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"), "Заголовок CORS не должен быть установлен")
		assert.True(t, nextCalled, "Запрос все равно должен пройти дальше (браузер сам заблокирует его)")
	})

	t.Run("Preflight запрос (OPTIONS)", func(t *testing.T) {
		nextCalled = false
		origin := "http://foodcourt.fun/"
		req := httptest.NewRequest(http.MethodOptions, "/api/any", nil)
		req.Header.Set("Origin", origin)
		rec := httptest.NewRecorder()

		mw.Handler(nextHandler).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Equal(t, origin, rec.Header().Get("Access-Control-Allow-Origin"))

		// Проверяем, что запрос прервался и не пошел дальше
		assert.False(t, nextCalled, "Для метода OPTIONS следующий хэндлер не должен вызываться")
	})

	t.Run("Отсутствие заголовка Origin", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest(http.MethodGet, "/api/any", nil)
		rec := httptest.NewRecorder()

		mw.Handler(nextHandler).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
		assert.True(t, nextCalled)
	})
}
