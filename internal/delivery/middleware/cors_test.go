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

	tests := []struct {
		name              string
		method            string
		origin            string
		expectedStatus    int
		expectedOrigin    string
		expectCredentials bool
		expectNextCalled  bool
	}{
		{
			name:              "Разрешенный Origin",
			method:            http.MethodPost,
			origin:            "http://localhost:2033",
			expectedStatus:    http.StatusOK,
			expectedOrigin:    "http://localhost:2033",
			expectCredentials: true,
			expectNextCalled:  true,
		},
		{
			name:             "Запрещенный Origin",
			method:           http.MethodPost,
			origin:           "http://hacker.com",
			expectedStatus:   http.StatusOK,
			expectedOrigin:   "",   // Заголовок CORS не должен быть установлен
			expectNextCalled: true, // Запрос все равно должен пройти дальше (браузер сам заблокирует его)
		},
		{
			name:             "Preflight запрос (OPTIONS)",
			method:           http.MethodOptions,
			origin:           "http://foodcourt.fun/",
			expectedStatus:   http.StatusNoContent,
			expectedOrigin:   "http://foodcourt.fun/",
			expectNextCalled: false, // Для метода OPTIONS следующий хэндлер не должен вызываться
		},
		{
			name:             "Отсутствие заголовка Origin",
			method:           http.MethodGet,
			origin:           "",
			expectedStatus:   http.StatusOK,
			expectedOrigin:   "",
			expectNextCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Флаг для проверки того, дошел ли запрос до следующего хэндлера
			nextCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(tt.method, "/api/any", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()

			mw.Handler(nextHandler).ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			assert.Equal(t, tt.expectedOrigin, rec.Header().Get("Access-Control-Allow-Origin"))

			if tt.expectCredentials {
				assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
			}

			assert.Equal(t, tt.expectNextCalled, nextCalled, "Проверка вызова следующего хэндлера")
		})
	}
}
