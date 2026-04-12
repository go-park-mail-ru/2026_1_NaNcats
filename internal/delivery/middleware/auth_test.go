package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain/mocks"
	ucMocks "github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAuthMiddleware_RequireAuth(t *testing.T) {
	type mockInit func(m *ucMocks.MockSessionUseCase)

	// Вспомогательный хэндлер, который стоит за мидлварью
	// Если он вызвался — значит мидлварь пропустила запрос дальше
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что в контексте действительно лежит правильный UserID
		id := r.Context().Value(UserIDKey)
		assert.IsType(t, 0, id)
		assert.NotEqual(t, 0, id)
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		hasCookie      bool
		cookieValue    string
		mockInit       mockInit
		expectedStatus int
		expectedBody   string // Для проверки текста ошибки
	}{
		{
			name:        "Успешная авторизация",
			hasCookie:   true,
			cookieValue: uuid.New().String(),
			mockInit: func(m *ucMocks.MockSessionUseCase) {
				// Ожидаем проверку сессии
				m.EXPECT().
					Check(gomock.Any(), gomock.Any()).
					Return(domain.Session{UserID: 1}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Ошибка: кука отсутствует",
			hasCookie:      false,
			mockInit:       func(m *ucMocks.MockSessionUseCase) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Session not found",
		},
		{
			name:           "Ошибка: невалидный UUID",
			hasCookie:      true,
			cookieValue:    "not-a-uuid",
			mockInit:       func(m *ucMocks.MockSessionUseCase) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:        "Ошибка: сессия протухла или не найдена",
			hasCookie:   true,
			cookieValue: uuid.New().String(),
			mockInit: func(m *ucMocks.MockSessionUseCase) {
				// Программируем UseCase вернуть ошибку
				m.EXPECT().
					Check(gomock.Any(), gomock.Any()).
					Return(domain.Session{}, domain.ErrSessionExpired)
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSessionUC := ucMocks.NewMockSessionUseCase(ctrl)
			tt.mockInit(mockSessionUC)

			nopLogger := mocks.NewNopLogger()

			mw := NewAuthMiddleware(mockSessionUC, nopLogger)

			req := httptest.NewRequest(http.MethodGet, "/api/profile", nil)
			if tt.hasCookie {
				req.AddCookie(&http.Cookie{Name: "session_id", Value: tt.cookieValue})
			}

			rec := httptest.NewRecorder()

			// Запускаем цепочку Middleware -> nextHandler
			// Через ServeHTTP запускаем выполнение хендлера
			mw.RequireAuth(nextHandler).ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				// Проверяем, что в теле ответа есть наше сообщение
				assert.Contains(t, rec.Body.String(), tt.expectedBody)
			}
		})
	}
}
