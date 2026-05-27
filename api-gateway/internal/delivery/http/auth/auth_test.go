package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/authclient"
	authMocks "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/authclient/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/userclient"
	userMocks "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/userclient/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	pbAuth "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/auth"
	pbUser "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/user"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func setupTestHandler(ctrl *gomock.Controller) (*AuthHandler, *authMocks.MockAuthClient, *userMocks.MockUserClient) {
	mockAuth := authMocks.NewMockAuthClient(ctrl)
	mockUser := userMocks.NewMockUserClient(ctrl)
	log := logger.NewNopLogger()
	val := validator.New()

	handler := NewAuthHandler(mockAuth, mockUser, log, val)
	return handler, mockAuth, mockUser
}

func TestAuthHandler_Register(t *testing.T) {
	type mockBehavior func(auth *authMocks.MockAuthClient, user *userMocks.MockUserClient)

	validReq := RegisterRequest{
		Name:     "Иван",
		Email:    "test@mail.ru",
		Password: "password123",
	}

	tests := []struct {
		name           string
		reqBody        interface{}
		headers        map[string]string
		mockBehavior   mockBehavior
		expectedStatus int
	}{
		{
			name:    "Успешная регистрация",
			reqBody: validReq,
			headers: map[string]string{"Idempotency-Key": "idem-123"},
			mockBehavior: func(auth *authMocks.MockAuthClient, user *userMocks.MockUserClient) {
				user.EXPECT().CreateUser(gomock.Any(), "Иван", "test@mail.ru", "password123", "idem-123").
					Return(int64(42), nil)

				auth.EXPECT().IssueSession(gomock.Any(), int64(42), "user", gomock.Any()).
					Return(&pbAuth.Session{
						Id:        "session-uuid",
						ExpiresAt: timestamppb.New(time.Now().Add(time.Hour)),
					}, nil)

				auth.EXPECT().SetCSRF(gomock.Any(), "session-uuid").
					Return("csrf-token-123", nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Ошибка: нет Idempotency-Key",
			reqBody:        validReq,
			headers:        map[string]string{}, // Пустые заголовки
			mockBehavior:   func(auth *authMocks.MockAuthClient, user *userMocks.MockUserClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Ошибка валидации (короткий пароль)",
			reqBody: RegisterRequest{
				Name:     "Иван",
				Email:    "test@mail.ru",
				Password: "short", // Меньше 8 символов
			},
			headers:        map[string]string{"Idempotency-Key": "idem-123"},
			mockBehavior:   func(auth *authMocks.MockAuthClient, user *userMocks.MockUserClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Ошибка: пользователь уже существует",
			reqBody: validReq,
			headers: map[string]string{"Idempotency-Key": "idem-123"},
			mockBehavior: func(auth *authMocks.MockAuthClient, user *userMocks.MockUserClient) {
				user.EXPECT().CreateUser(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(int64(0), userclient.ErrEmailAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler, mockAuth, mockUser := setupTestHandler(ctrl)
			tt.mockBehavior(mockAuth, mockUser)

			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))

			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			w := httptest.NewRecorder()
			handler.Register(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusCreated {
				assert.NotEmpty(t, w.Header().Get("Set-Cookie")) // Проверяем установку куки
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	type mockBehavior func(auth *authMocks.MockAuthClient, user *userMocks.MockUserClient)

	validReq := LoginRequest{
		Login:    "test@mail.ru",
		Password: "password123",
	}

	tests := []struct {
		name           string
		reqBody        interface{}
		mockBehavior   mockBehavior
		expectedStatus int
	}{
		{
			name:    "Успешная авторизация",
			reqBody: validReq,
			mockBehavior: func(auth *authMocks.MockAuthClient, user *userMocks.MockUserClient) {
				auth.EXPECT().Login(gomock.Any(), "test@mail.ru", "password123", gomock.Any()).
					Return(&pbAuth.Session{
						Id:        "session-uuid",
						UserId:    42,
						ExpiresAt: timestamppb.New(time.Now().Add(time.Hour)),
					}, nil)

				// Предполагаем, что GetByID возвращает профиль пользователя
				user.EXPECT().GetByID(gomock.Any(), int64(42)).
					Return(&pbUser.User{
						Name:      "Иван",
						Email:     "test@mail.ru",
						AvatarUrl: "url",
					}, nil)

				auth.EXPECT().SetCSRF(gomock.Any(), "session-uuid").
					Return("csrf-token-123", nil)

				user.EXPECT().GetUserProfile(gomock.Any(), int64(42)).
					Return(nil, &pbUser.ClientProfile{StreakCount: 3}, nil).AnyTimes()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Ошибка: неверные учетные данные",
			reqBody: validReq,
			mockBehavior: func(auth *authMocks.MockAuthClient, user *userMocks.MockUserClient) {
				auth.EXPECT().Login(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, authclient.ErrInvalidCredentials)
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler, mockAuth, mockUser := setupTestHandler(ctrl)
			tt.mockBehavior(mockAuth, mockUser)

			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.Login(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, mockAuth, _ := setupTestHandler(ctrl)

	t.Run("Успешный логаут с удалением куки", func(t *testing.T) {
		mockAuth.EXPECT().Logout(gomock.Any(), "session-uuid").Return(nil)

		req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "session-uuid"})
		w := httptest.NewRecorder()

		handler.Logout(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Проверяем, что кука была перезаписана на пустую с истекшим сроком годности
		cookieStr := w.Header().Get("Set-Cookie")
		require.NotEmpty(t, cookieStr)
		assert.Contains(t, cookieStr, "session_id=")
		assert.Contains(t, cookieStr, "1970") // Признак сброшенного времени (Unix 0)
	})

	t.Run("Логаут без куки (не падает)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
		w := httptest.NewRecorder()

		handler.Logout(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// Вспомогательная функция для инжекта UserID в контекст
func withUserIDContext(req *http.Request, userID int64) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	return req.WithContext(ctx)
}

func TestAuthHandler_GetMe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, _, mockUser := setupTestHandler(ctrl)

	t.Run("Успешное получение профиля", func(t *testing.T) {
		mockUser.EXPECT().GetByID(gomock.Any(), int64(42)).
			Return(&pbUser.User{
				Name:      "Иван",
				Email:     "test@mail.ru",
				AvatarUrl: "img.png",
			}, nil)
		mockUser.EXPECT().GetUserProfile(gomock.Any(), int64(42)).
			Return(nil, &pbUser.ClientProfile{StreakCount: 3}, nil).AnyTimes()

		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
		req = withUserIDContext(req, 42) // Имитируем успешную работу middleware
		w := httptest.NewRecorder()

		handler.GetMe(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Ошибка: неавторизован (нет в контексте)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
		// Не добавляем UserID в контекст
		w := httptest.NewRecorder()

		handler.GetMe(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestAuthHandler_GetCSRF(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, mockAuth, _ := setupTestHandler(ctrl)

	t.Run("Успешное получение CSRF", func(t *testing.T) {
		mockAuth.EXPECT().GetCSRF(gomock.Any(), "valid-session").
			Return("new-csrf-token", nil)

		req := httptest.NewRequest(http.MethodGet, "/csrf", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session"})
		w := httptest.NewRecorder()

		handler.GetCSRF(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp CSRFResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "new-csrf-token", resp.CSRFToken)
	})

	t.Run("Нет куки сессии (возвращает 200 и сообщение)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/csrf", nil)
		w := httptest.NewRecorder()

		handler.GetCSRF(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "no session")
	})

	t.Run("Сессия невалидна / истекла", func(t *testing.T) {
		mockAuth.EXPECT().GetCSRF(gomock.Any(), "invalid-session").
			Return("", authclient.ErrSessionNotFound)

		req := httptest.NewRequest(http.MethodGet, "/csrf", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "invalid-session"})
		w := httptest.NewRecorder()

		handler.GetCSRF(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
