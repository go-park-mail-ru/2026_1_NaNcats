package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/authclient"
	authMocks "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/authclient/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func makeAuthMW(t *testing.T) (*AuthMiddleware, *authMocks.MockAuthClient, *gomock.Controller) {
	t.Helper()
	ctrl := gomock.NewController(t)
	ac := authMocks.NewMockAuthClient(ctrl)
	mw := NewAuthMiddleware(ac, logger.NewNopLogger())
	return mw, ac, ctrl
}

// nextOK — простой пробник, который запоминает, что был вызван.
type probe struct{ called bool }

func TestAuthMiddleware_RequireAuth(t *testing.T) {
	tests := []struct {
		name           string
		withCookie     bool
		cookieValue    string
		setupMock      func(ac *authMocks.MockAuthClient)
		expectedStatus int
		expectNextCall bool
	}{
		{
			name:           "Без cookie — 401",
			withCookie:     false,
			setupMock:      func(ac *authMocks.MockAuthClient) {},
			expectedStatus: http.StatusUnauthorized,
			expectNextCall: false,
		},
		{
			name:        "Сессия не найдена — 401",
			withCookie:  true,
			cookieValue: "stale-session",
			setupMock: func(ac *authMocks.MockAuthClient) {
				ac.EXPECT().CheckSession(gomock.Any(), "stale-session").
					Return(int64(0), "", authclient.ErrSessionNotFound)
			},
			expectedStatus: http.StatusUnauthorized,
			expectNextCall: false,
		},
		{
			name:        "Внутренняя ошибка auth-сервиса — 500",
			withCookie:  true,
			cookieValue: "session",
			setupMock: func(ac *authMocks.MockAuthClient) {
				ac.EXPECT().CheckSession(gomock.Any(), "session").
					Return(int64(0), "", errors.New("auth down"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectNextCall: false,
		},
		{
			name:        "Успешная авторизация — пропуск дальше + контекст",
			withCookie:  true,
			cookieValue: "session-ok",
			setupMock: func(ac *authMocks.MockAuthClient) {
				ac.EXPECT().CheckSession(gomock.Any(), "session-ok").
					Return(int64(42), "owner", nil)
			},
			expectedStatus: http.StatusOK,
			expectNextCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw, ac, ctrl := makeAuthMW(t)
			defer ctrl.Finish()
			tt.setupMock(ac)

			p := &probe{}
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				p.called = true
				if tt.expectNextCall {
					id, ok := GetUserID(r.Context())
					assert.True(t, ok)
					assert.EqualValues(t, 42, id)
					role, ok := GetUserRole(r.Context())
					assert.True(t, ok)
					assert.Equal(t, "owner", role)
				}
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.withCookie {
				req.AddCookie(&http.Cookie{Name: "session_id", Value: tt.cookieValue})
			}
			w := httptest.NewRecorder()
			mw.RequireAuth(next).ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectNextCall, p.called)
		})
	}
}

func TestAuthMiddleware_OptionalAuth(t *testing.T) {
	t.Run("Без cookie — пропуск как гость", func(t *testing.T) {
		mw, _, ctrl := makeAuthMW(t)
		defer ctrl.Finish()

		var seenUser bool
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, seenUser = GetUserID(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		mw.OptionalAuth(next).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.False(t, seenUser, "контекст должен быть пустым для гостя")
	})

	t.Run("С невалидной сессией — пропуск как гость", func(t *testing.T) {
		mw, ac, ctrl := makeAuthMW(t)
		defer ctrl.Finish()
		ac.EXPECT().CheckSession(gomock.Any(), "x").Return(int64(0), "", errors.New("err"))

		var seenUser bool
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, seenUser = GetUserID(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "x"})
		w := httptest.NewRecorder()
		mw.OptionalAuth(next).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.False(t, seenUser)
	})

	t.Run("С валидной сессией — userID и role в контексте", func(t *testing.T) {
		mw, ac, ctrl := makeAuthMW(t)
		defer ctrl.Finish()
		ac.EXPECT().CheckSession(gomock.Any(), "ok").Return(int64(7), "client", nil)

		var gotID int64
		var gotRole string
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotID, _ = GetUserID(r.Context())
			gotRole, _ = GetUserRole(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "ok"})
		w := httptest.NewRecorder()
		mw.OptionalAuth(next).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.EqualValues(t, 7, gotID)
		assert.Equal(t, "client", gotRole)
	})
}

func TestAuthMiddleware_RequireRole(t *testing.T) {
	mw, _, ctrl := makeAuthMW(t)
	defer ctrl.Finish()
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	t.Run("Нет роли в контексте — 500 (предполагается что RequireAuth вызывался раньше)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		mw.RequireRole("owner")(next).ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Роль есть и совпадает — пропуск", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := context.WithValue(req.Context(), RoleKey, "owner")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()
		mw.RequireRole("owner", "admin")(next).ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Роль не входит в список — 403", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := context.WithValue(req.Context(), RoleKey, "client")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()
		mw.RequireRole("owner")(next).ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestContextKeys(t *testing.T) {
	t.Run("GetUserID — пустой контекст", func(t *testing.T) {
		_, ok := GetUserID(context.Background())
		assert.False(t, ok)
	})
	t.Run("GetUserID — значение в контексте", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), UserIDKey, int64(99))
		id, ok := GetUserID(ctx)
		assert.True(t, ok)
		assert.EqualValues(t, 99, id)
	})
	t.Run("GetUserRole — пустая строка считается отсутствием роли", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), RoleKey, "")
		_, ok := GetUserRole(ctx)
		assert.False(t, ok)
	})
	t.Run("GetUserRole — нормальная роль", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), RoleKey, "admin")
		role, ok := GetUserRole(ctx)
		assert.True(t, ok)
		assert.Equal(t, "admin", role)
	})
}

func TestRequestIDMiddleware(t *testing.T) {
	mw := NewRequestIDMiddleware()
	assert.NotNil(t, mw)

	t.Run("Генерация UUID, если заголовок отсутствует", func(t *testing.T) {
		var ctxReqID string
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, _ := r.Context().Value(common.RequestIDKey).(string)
			ctxReqID = id
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		mw.Handler(next).ServeHTTP(w, req)

		assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
		assert.NotEmpty(t, ctxReqID, "request id должен пробрасываться в контекст")
		assert.Equal(t, w.Header().Get("X-Request-ID"), ctxReqID)
	})

	t.Run("Сохраняет переданный X-Request-ID", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Request-ID", "trace-123")
		w := httptest.NewRecorder()
		mw.Handler(next).ServeHTTP(w, req)
		assert.Equal(t, "trace-123", w.Header().Get("X-Request-ID"))
	})
}

func TestLoggingMiddleware(t *testing.T) {
	mw := NewLoggingMiddleware(logger.NewNopLogger())
	assert.NotNil(t, mw)

	t.Run("Передаёт запрос дальше и не ломает 200", func(t *testing.T) {
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("ok"))
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		mw.Handler(next).ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "ok", w.Body.String())
	})

	t.Run("Status code и size правильно фиксируются обёрткой", func(t *testing.T) {
		rec := httptest.NewRecorder()
		wrapped := &responseWriterWrapper{ResponseWriter: rec, statusCode: http.StatusOK}
		wrapped.WriteHeader(http.StatusTeapot)
		n, err := wrapped.Write([]byte("hello"))
		assert.NoError(t, err)
		assert.Equal(t, 5, n)
		assert.Equal(t, http.StatusTeapot, wrapped.statusCode)
		assert.Equal(t, 5, wrapped.size)
	})

	t.Run("Hijack без поддержки — ошибка", func(t *testing.T) {
		// httptest.ResponseRecorder не имплементирует Hijacker
		rec := httptest.NewRecorder()
		wrapped := &responseWriterWrapper{ResponseWriter: rec}
		_, _, err := wrapped.Hijack()
		assert.Error(t, err)
		assert.Contains(t, strings.ToLower(err.Error()), "hijacker")
	})
}
