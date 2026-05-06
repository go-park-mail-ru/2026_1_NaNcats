package middleware

import (
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/authclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/response"
)

type CSRFMiddleware struct {
	authClient authclient.AuthClient
	logger     logger.Logger
}

func NewCSRFMiddleware(ac authclient.AuthClient, l logger.Logger) *CSRFMiddleware {
	return &CSRFMiddleware{
		authClient: ac,
		logger:     l,
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

		clientToken := r.Header.Get("X-CSRF-Token")
		if clientToken == "" {
			response.Error(w, http.StatusForbidden, "CSRF token missing")
			return
		}

		expectedToken, err := m.authClient.GetCSRF(ctx, cookie.Value)
		if err != nil {
			if errors.Is(err, authclient.ErrSessionNotFound) {
				l.Warn("csrf: session not found in auth service", logger.String("session_id", cookie.Value))
				response.Error(w, http.StatusForbidden, "Invalid or expired session")
				return
			}
			l.Error("csrf: validation failed due to internal error", err)
			response.Error(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		if clientToken != expectedToken {
			l.Warn("csrf: token mismatch", logger.String("session_id", cookie.Value))
			response.Error(w, http.StatusForbidden, "CSRF token mismatch")
			return
		}

		next.ServeHTTP(w, r)
	})
}
