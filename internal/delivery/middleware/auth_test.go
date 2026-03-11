package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	ucMocks "github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAuthMiddleware_RequireAuth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSessionUC := ucMocks.NewMockSessionUseCase(ctrl)
	mw := NewAuthMiddleware(mockSessionUC)

	// Вспомогательный хэндлер, который стоит за мидлварью
	// Если он вызвался — значит мидлварь пропустила запрос дальше
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что в контексте действительно лежит правильный UserID
		uid := r.Context().Value(UserIDKey)
		assert.NotNil(t, uid)
		w.WriteHeader(http.StatusOK)
	})

	t.Run("Успешная авторизация", func(t *testing.T) {
		sessID := uuid.New()
		userID := uuid.New()

		req := httptest.NewRequest(http.MethodGet, "/api/profile", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessID.String()})
		rec := httptest.NewRecorder()

		// Ожидаем проверку сессии
		mockSessionUC.EXPECT().
			Check(gomock.Any(), sessID).
			Return(userID, nil)

		// Запускаем цепочку Middleware -> nextHandler
		// Через ServeHTTP запускаем выполнение хендлера
		mw.RequireAuth(nextHandler).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Ошибка: кука отсутствует", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/profile", nil)
		rec := httptest.NewRecorder()

		// Запускаем
		mw.RequireAuth(nextHandler).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		// Проверяем, что в теле ответа есть наше сообщение
		assert.Contains(t, rec.Body.String(), "Session not found")
	})

	t.Run("Ошибка: невалидный UUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/profile", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "not-a-uuid"})
		rec := httptest.NewRecorder()

		mw.RequireAuth(nextHandler).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Ошибка: сессия протухла или не найдена", func(t *testing.T) {
		sessID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/api/profile", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessID.String()})
		rec := httptest.NewRecorder()

		// Программируем UseCase вернуть ошибку
		mockSessionUC.EXPECT().
			Check(gomock.Any(), sessID).
			Return(uuid.Nil, domain.ErrSessionExpired)

		mw.RequireAuth(nextHandler).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
