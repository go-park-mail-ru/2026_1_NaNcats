package middleware

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	auth "github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/auth"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/response"
	"github.com/google/uuid"
)

type CSRFMiddleware struct {
	sessionUC auth.SessionUseCase
	logger    domain.Logger
}

func NewCSRFMiddleware(suc auth.SessionUseCase, l domain.Logger) *CSRFMiddleware {
	return &CSRFMiddleware{
		sessionUC: suc,
		logger:    l,
	}
}

func (m *CSRFMiddleware) Check(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := m.logger.WithContext(ctx)

		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie("session_id")
		if err != nil {
			l.Warn("csrf: session cookie not found")
			response.Error(w, http.StatusForbidden, "Session required for CSRF check")
			return
		}

		sessionID, err := uuid.Parse(cookie.Value)
		if err != nil {
			l.Warn("csrf: invalid session token format", domain.String("token_value", cookie.Value))
			response.Error(w, http.StatusUnauthorized, "Invalid session")
			return
		}

		clientToken := r.Header.Get("X-CSRF-Token")
		if clientToken == "" {
			response.Error(w, http.StatusForbidden, "CSRF token missing")
			return
		}

		expectedToken, err := m.sessionUC.GetCSRF(ctx, sessionID)
		if err != nil {
			l.Error("csrf: validation failed", err, domain.String("session_id", sessionID.String()))
			response.Error(w, http.StatusForbidden, "Invalid or expired CSRF token")
			return
		}

		if clientToken != expectedToken {
			l.Warn("csrf: token mismatch", domain.String("session_id", sessionID.String()))
			response.Error(w, http.StatusForbidden, "CSRF token mismatch")
			return
		}

		next.ServeHTTP(w, r)
	})
}
