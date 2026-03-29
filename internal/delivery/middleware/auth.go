package middleware

import (
	"context"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/response"
	"github.com/google/uuid"
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
			return
		}

		ctx := r.Context()

		session, err := m.sessionUC.Check(ctx, sessionID)
		if err != nil {
			response.Error(w, http.StatusUnauthorized, "Invalid or expired session")
			return
		}

		currentUserAgent := r.UserAgent()
		if session.UserAgent != currentUserAgent {
			m.sessionUC.Destroy(r.Context(), sessionID)
			return
		}

		// добавляем к контексту ctx ключ UserIDKey со значением userID
		ctxWithUser := context.WithValue(ctx, UserIDKey, session.UserID)

		// отдаем обработать запрос дальше
		next.ServeHTTP(w, r.WithContext(ctxWithUser))
	})
}
