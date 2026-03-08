package middleware

import (
	"context"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/response"
	"github.com/google/uuid"
)

// уникальный тип ключа контекста:
// нужен, чтобы предотвратить коллизии с другими пакетами в контексте
type contextKey string

const (
	// ключ ID пользователя для контекста
	UserIDKey contextKey = "userID"
)

// мидлвар авторизации:
// нужен для защиты приватных эндпоинтов от forbidden/unauthorized сессий
type AuthMiddleware struct {
	sessionUC usecase.SessionUseCase
}

func NewAuthMiddleware(suc usecase.SessionUseCase) *AuthMiddleware {
	return &AuthMiddleware{
		sessionUC: suc,
	}
}

// защита от unauthorized сессий
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			response.Error(w, http.StatusUnauthorized, "Session not found")
			return
		}

		sessionID, err := uuid.Parse(cookie.Value)
		if err != nil {
			response.Error(w, http.StatusUnauthorized, "Invalid session")
		}

		ctx := r.Context()

		userID, err := m.sessionUC.Check(ctx, sessionID)
		if err != nil {
			response.Error(w, http.StatusUnauthorized, "Invalid or expired session")
			return
		}

		// добавляем к контексту ctx ключ UserIDKey со значение userID
		ctxWithUser := context.WithValue(ctx, UserIDKey, userID)

		// отдаем обработать запрос дальше
		next.ServeHTTP(w, r.WithContext(ctxWithUser))
	})
}
