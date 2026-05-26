package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/authclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/response"
)

// мидлвар авторизации:
// нужен для защиты приватных эндпоинтов от forbidden/unauthorized сессий
type AuthMiddleware struct {
	authClient authclient.AuthClient
	logger     logger.Logger
}

func NewAuthMiddleware(ac authclient.AuthClient, l logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authClient: ac,
		logger:     l,
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

		userID, role, err := m.authClient.CheckSession(ctx, cookie.Value)
		if err != nil {
			if errors.Is(err, authclient.ErrSessionNotFound) {
				l.Info("auth mw: unauthorized access attempt or expired session",
					logger.String("session_id", cookie.Value),
				)
				response.Error(w, http.StatusUnauthorized, "Invalid or expired session")
				return
			}

			l.Error("auth mw: auth service critical failure", err,
				logger.String("session_id", cookie.Value),
			)
			response.Error(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		ctxWithUser := context.WithValue(ctx, UserIDKey, userID)
		ctxWithUser = context.WithValue(ctxWithUser, RoleKey, role)

		// Отдаем обработать запрос дальше
		next.ServeHTTP(w, r.WithContext(ctxWithUser))
	})
}

// OptionalAuth прокидывает userID/role в контекст, если в куке валидная
// сессия; иначе пропускает запрос дальше как анонимный (без 401). Используется
// эндпоинтами, у которых поведение различается для гостя и авторизованного.
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		cookie, err := r.Cookie("session_id")
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		userID, role, err := m.authClient.CheckSession(ctx, cookie.Value)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctxWithUser := context.WithValue(ctx, UserIDKey, userID)
		ctxWithUser = context.WithValue(ctxWithUser, RoleKey, role)
		next.ServeHTTP(w, r.WithContext(ctxWithUser))
	})
}

func (m *AuthMiddleware) RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			userRole, ok := GetUserRole(ctx)
			if !ok {
				m.logger.WithContext(ctx).Error("auth mw: role missing from context. ensure RequireAuth is called before RequireRole", nil)
				response.Error(w, http.StatusInternalServerError, "Internal server error")
				return
			}

			isAllowed := false
			for _, role := range allowedRoles {
				if userRole == role {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
				m.logger.WithContext(ctx).Warn("auth mw: access denied for role",
					logger.String("user_role", userRole),
					logger.Any("required_roles", allowedRoles),
				)
				response.Error(w, http.StatusForbidden, "Access denied")
				return
			}

			next.ServeHTTP(w, r)
		})

	}
}
