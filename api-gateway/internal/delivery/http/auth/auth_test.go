package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/authclient"
	authMocks "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/authclient/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/userclient"
	userMocks "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/userclient/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Вспомогательные структуры для имитации ответов gRPC клиентов,
// подставь реальные типы из твоих proto-файлов, если они отличаются.
type mockUserResp struct {
	Id        int64
	Name      string
	Email     string
	AvatarUrl string
}

type mockSessionResp struct {
	Id        string
	UserId    int64
	ExpiresAt *timestamppb.Timestamp
}

func TestAuthHandler_Register(t *testing.T) {
	type mockInit func(authMock *authMocks.MockAuthClient, userMock *userMocks.MockUserClient)

	val := validator.New()
	nopLogger := logger.NewNopLogger()

	validReq := RegisterRequest{
		Name:     "Иван",
		Email:    "test@mail.ru",
		Password: "password123",
	}

	tests := []struct {
		name           string
		reqBody        any
		headers        map[string]string
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:    "Успешная регистрация",
			reqBody: validReq,
			headers: map[string]string{"Idempotency-Key": "idem-123"},
			mockInit: func(authMock *authMocks.MockAuthClient, userMock *userMocks.MockUserClient) {
				userResp := &mockUserResp{Id: 1, Name: "Иван", Email: "test@mail.ru"}
				sessionResp := &mockSessionResp{Id: "sess-uuid", UserId: 1, ExpiresAt: timestamppb.Now()}

				userMock.EXPECT().CreateUser(gomock.Any(), "Иван", "test@mail.ru", "password123", "idem-123").
					Return(userResp, nil)
				authMock.EXPECT().IssueSession(gomock.Any(), userResp, "user", gomock.Any()).
					Return(sessionResp, nil)
				authMock.EXPECT().SetCSRF(gomock.Any(), "sess-uuid").
					Return("csrf-token-123", nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Ошибка: нет Idempotency-Key",
			reqBody:        validReq,
			headers:        map[string]string{}, // Пустые заголовки
			mockInit:       func(a *authMocks.MockAuthClient, u *userMocks.MockUserClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Ошибка валидации (короткий пароль)",
			reqBody: RegisterRequest{
				Name:     "Иван",
				Email:    "test@mail.ru",
				Password: "123", // min=8
			},
			headers:        map[string]string{"Idempotency-Key": "idem-123"},
			mockInit:       func(a *authMocks.MockAuthClient, u *userMocks.MockUserClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Ошибка: email уже существует",
			reqBody: validReq,
			headers: map[string]string{"Idempotency-Key": "idem-123"},
			mockInit: func(authMock *authMocks.MockAuthClient, userMock *userMocks.MockUserClient) {
				userMock.EXPECT().CreateUser(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, userclient.ErrEmailAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authMock := authMocks.NewMockAuthClient(ctrl)
			userMock := userMocks.NewMockUserClient(ctrl)
			tt.mockInit(authMock, userMock)

			handler := NewAuthHandler(authMock, userMock, nopLogger, val)

			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			rr := httptest.NewRecorder()
			handler.Register(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusCreated {
				assert.Contains(t, rr.Header().Get("Set-Cookie"), "session_id")
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	type mockInit func(authMock *authMocks.MockAuthClient, userMock *userMocks.MockUserClient)

	val := validator.New()
	nopLogger := logger.NewNopLogger()

	validReq := LoginRequest{
		Login:    "test@mail.ru",
		Password: "password123",
	}

	tests := []struct {
		name           string
		reqBody        any
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:    "Успешный вход",
			reqBody: validReq,
			mockInit: func(authMock *authMocks.MockAuthClient, userMock *userMocks.MockUserClient) {
				sessionResp := &mockSessionResp{Id: "sess-uuid", UserId: 1, ExpiresAt: timestamppb.Now()}
				userResp := &mockUserResp{Name: "Иван", Email: "test@mail.ru"}

				authMock.EXPECT().Login(gomock.Any(), "test@mail.ru", "password123", gomock.Any()).
					Return(sessionResp, nil)
				userMock.EXPECT().GetByID(gomock.Any(), int64(1)).
					Return(userResp, nil)
				authMock.EXPECT().SetCSRF(gomock.Any(), "sess-uuid").
					Return("csrf-token-123", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Ошибка: неверные учетные данные",
			reqBody: validReq,
			mockInit: func(authMock *authMocks.MockAuthClient, userMock *userMocks.MockUserClient) {
				authMock.EXPECT().Login(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, authclient.ErrInvalidCredentials)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Ошибка валидации",
			reqBody: LoginRequest{
				Login:    "test@mail.ru",
				Password: "123", // min=8
			},
			mockInit:       func(a *authMocks.MockAuthClient, u *userMocks.MockUserClient) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authMock := authMocks.NewMockAuthClient(ctrl)
			userMock := userMocks.NewMockUserClient(ctrl)
			tt.mockInit(authMock, userMock)

			handler := NewAuthHandler(authMock, userMock, nopLogger, val)

			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.Login(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				assert.Contains(t, rr.Header().Get("Set-Cookie"), "session_id")
			}
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authMock := authMocks.NewMockAuthClient(ctrl)
	handler := NewAuthHandler(authMock, nil, logger.NewNopLogger(), validator.New())

	t.Run("Успешный выход с кукой", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session"})

		// Ожидаем вызов Logout в gRPC клиенте
		authMock.EXPECT().Logout(gomock.Any(), "valid-session").Return(nil)

		rr := httptest.NewRecorder()
		handler.Logout(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		// Проверяем, что кука затирается (MaxAge < 0 или пустая)
		assert.Contains(t, rr.Header().Get("Set-Cookie"), "session_id=")
	})

	t.Run("Успешный выход без куки", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)

		// gRPC Logout не должен вызываться
		rr := httptest.NewRecorder()
		handler.Logout(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestAuthHandler_GetMe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userMock := userMocks.NewMockUserClient(ctrl)
	handler := NewAuthHandler(nil, userMock, logger.NewNopLogger(), validator.New())

	t.Run("Успешное получение профиля", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)

		// Эмулируем работу мидлвари, кладем UserID в контекст
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, int64(42))
		req = req.WithContext(ctx)

		userMock.EXPECT().GetByID(gomock.Any(), int64(42)).
			Return(&mockUserResp{Name: "Иван", Email: "ivan@test.ru"}, nil)

		rr := httptest.NewRecorder()
		handler.GetMe(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp LoginResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "Иван", resp.Name)
	})

	t.Run("Неавторизован (нет UserID в контексте)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)

		rr := httptest.NewRecorder()
		handler.GetMe(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

func TestAuthHandler_GetCSRF(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authMock := authMocks.NewMockAuthClient(ctrl)
	handler := NewAuthHandler(authMock, nil, logger.NewNopLogger(), validator.New())

	t.Run("Успешное получение токена", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/csrf", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "sess-123"})

		authMock.EXPECT().GetCSRF(gomock.Any(), "sess-123").Return("csrf-abc", nil)

		rr := httptest.NewRecorder()
		handler.GetCSRF(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp CSRFResponse
		json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Equal(t, "csrf-abc", resp.CSRFToken)
	})

	t.Run("Нет сессии (сообщение, код 200)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/csrf", nil)

		rr := httptest.NewRecorder()
		handler.GetCSRF(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp CSRFResponse
		json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Equal(t, "no session", resp.Message)
	})

	t.Run("Сессия невалидна (протухла)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/csrf", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "sess-expired"})

		authMock.EXPECT().GetCSRF(gomock.Any(), "sess-expired").Return("", authclient.ErrSessionNotFound)

		rr := httptest.NewRecorder()
		handler.GetCSRF(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}
