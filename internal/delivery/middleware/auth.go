package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	auth "github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/auth"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/response"
	"github.com/google/uuid"
)

// мидлвар авторизации:
// нужен для защиты приватных эндпоинтов от forbidden/unauthorized сессий
type AuthMiddleware struct {
	sessionUC auth.SessionUseCase
	logger    domain.Logger
}

func NewAuthMiddleware(suc auth.SessionUseCase, logger domain.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		sessionUC: suc,
		logger:    logger,
	}
}

// защита от unauthorized сессий
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := m.logger.WithContext(ctx)

		cookie, err := r.Cookie("session_id")
		if err != nil {
			response.Error(w, http.StatusUnauthorized, "Session not found")
			return
		}

		sessionID, err := uuid.Parse(cookie.Value)
		if err != nil {
			l.Warn("logout: invalid session token format", domain.String("token_value", cookie.Value))
			response.Error(w, http.StatusUnauthorized, "Invalid session")
			return
		}

		session, err := m.sessionUC.Check(ctx, sessionID)
		if err != nil {
			if errors.Is(err, domain.ErrSessionNotFound) || errors.Is(err, domain.ErrSessionExpired) {
				l.Info("auth: unauthorized access attempt",
					domain.String("session_id", sessionID.String()),
					domain.String("reason", err.Error()),
				)
				response.Error(w, http.StatusUnauthorized, "Invalid or expired session")
				return
			}

			l.Error("auth: session service critical failure", err,
				domain.String("session_id", sessionID.String()),
			)

			response.Error(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		currentUserAgent := r.UserAgent()
		if session.UserAgent != currentUserAgent {
			l.Warn("session user-agent mismatch - potential hijacking attempt",
				domain.String("session_id", sessionID.String()),
				domain.Int("user_id", session.UserID),
				domain.String("expected", session.UserAgent),
				domain.String("actual", currentUserAgent),
				domain.String("ip", r.RemoteAddr),
			)
			m.sessionUC.Destroy(r.Context(), sessionID)
			return
		}

		// добавляем к контексту ctx ключ UserIDKey со значением userID
		ctxWithUser := context.WithValue(ctx, UserIDKey, session.UserID)

		// отдаем обработать запрос дальше
		next.ServeHTTP(w, r.WithContext(ctxWithUser))
	})
}
