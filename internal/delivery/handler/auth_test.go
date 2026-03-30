package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	domainMocks "github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain/mocks"
	ucMocks "github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/response"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAuthHandler_Register(t *testing.T) {
	type mockInit func(m *ucMocks.MockAuthUseCase)

	tests := []struct {
		name           string
		inputBody      any // interface{} для поддержки и структур, и кривых строк
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name: "Успешная регистрация",
			inputBody: RegisterRequest{
				Name:     "Ivan",
				Email:    "test@mail.ru",
				Password: "password123",
			},
			mockInit: func(m *ucMocks.MockAuthUseCase) {
				mockUser := domain.User{ID: 1, Name: "Ivan", Email: "test@mail.ru"}
				mockSess := domain.Session{ID: uuid.New(), ExpiresAt: time.Now().Add(time.Hour)}
				m.EXPECT().
					Register(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(mockUser, mockSess, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Неверный JSON",
			inputBody:      "invalid-json",
			mockInit:       func(m *ucMocks.MockAuthUseCase) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuthUC := ucMocks.NewMockAuthUseCase(ctrl)
			testCase.mockInit(mockAuthUC)

			mockLogger := domainMocks.NewMockLogger(ctrl)

			mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
			mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
			mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			authHandler := NewAuthHandler(mockAuthUC, mockLogger)

			var buf bytes.Buffer
			if s, ok := testCase.inputBody.(string); ok {
				buf.WriteString(s)
			} else {
				err := json.NewEncoder(&buf).Encode(testCase.inputBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/auth/register", &buf)
			rec := httptest.NewRecorder()

			authHandler.Register(rec, req)

			assert.Equal(t, testCase.expectedStatus, rec.Code)

			if testCase.expectedStatus == http.StatusCreated {
				assert.Contains(t, rec.Header().Get("Set-Cookie"), "session_id")

				var resp RegisterResponse
				err := json.NewDecoder(rec.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.NotEmpty(t, resp.Name)
			}
		})
	}
}

func TestAuthHandler_GetMe(t *testing.T) {
	type mockInit func(m *ucMocks.MockAuthUseCase)

	tests := []struct {
		name             string
		userID           any // int или nil
		mockInit         mockInit
		expectedStatus   int
		expectedJSONCode int
	}{
		{
			name:   "Успешный запуск",
			userID: 1,
			mockInit: func(m *ucMocks.MockAuthUseCase) {
				m.EXPECT().
					GetProfile(gomock.Any(), gomock.Any()).
					Return(domain.User{Name: "Ivan"}, nil)
			},
			expectedStatus:   http.StatusOK,
			expectedJSONCode: 0, // если не 5xx, то ничего не ожидаем
		},
		{
			name:             "Нет юзера",
			userID:           nil,
			mockInit:         func(m *ucMocks.MockAuthUseCase) {},
			expectedStatus:   http.StatusOK,
			expectedJSONCode: http.StatusInternalServerError,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuthUC := ucMocks.NewMockAuthUseCase(ctrl)
			testCase.mockInit(mockAuthUC)

			mockLogger := domainMocks.NewMockLogger(ctrl)

			mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
			mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
			mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			authHandler := NewAuthHandler(mockAuthUC, mockLogger)

			req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
			if testCase.userID != nil {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, testCase.userID)
				req = req.WithContext(ctx)
			}

			rec := httptest.NewRecorder()

			authHandler.GetMe(rec, req)

			assert.Equal(t, testCase.expectedStatus, rec.Code)

			if testCase.expectedJSONCode != 0 {
				var resp response.ErrorResponse
				err := json.NewDecoder(rec.Body).Decode(&resp)

				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedJSONCode, resp.Code)
			} else if testCase.expectedStatus == http.StatusOK {
				assert.Contains(t, rec.Body.String(), "Ivan")
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	type mockInit func(m *ucMocks.MockAuthUseCase)

	tests := []struct {
		name           string
		inputBody      any
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name: "Успешный вход",
			inputBody: LoginRequest{
				Login:    "test@gmail.com",
				Password: "aboba",
			},
			mockInit: func(m *ucMocks.MockAuthUseCase) {
				mockUser := domain.User{ID: 1, Name: "Ivan", Email: "test@mail.ru"}
				mockSess := domain.Session{ID: uuid.New(), ExpiresAt: time.Now().Add(time.Hour)}

				m.EXPECT().
					Login(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(mockUser, mockSess, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Неверный JSON",
			inputBody:      "invalid JSON",
			mockInit:       func(m *ucMocks.MockAuthUseCase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Неверный пароль",
			inputBody: LoginRequest{
				Login:    "test@mail.ru",
				Password: "wrong",
			},
			mockInit: func(m *ucMocks.MockAuthUseCase) {
				// Программируем UseCase вернуть ошибку
				m.EXPECT().
					Login(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(domain.User{}, domain.Session{}, domain.ErrInvalidCredentials)
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuthUC := ucMocks.NewMockAuthUseCase(ctrl)
			testCase.mockInit(mockAuthUC)

			mockLogger := domainMocks.NewMockLogger(ctrl)

			mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
			mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
			mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			authHandler := NewAuthHandler(mockAuthUC, mockLogger)

			var buf bytes.Buffer
			if s, ok := testCase.inputBody.(string); ok {
				buf.WriteString(s)
			} else {
				err := json.NewEncoder(&buf).Encode(testCase.inputBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", &buf)
			rec := httptest.NewRecorder()

			authHandler.Login(rec, req)

			assert.Equal(t, testCase.expectedStatus, rec.Code)

			if testCase.expectedStatus == http.StatusOK {
				assert.Contains(t, rec.Header().Get("Set-Cookie"), "session_id")

				var resp LoginResponse
				err := json.NewDecoder(rec.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.NotEmpty(t, resp.Name)
			}
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	type mockInit func(m *ucMocks.MockAuthUseCase)

	tests := []struct {
		name           string
		cookieValue    string // Значение куки session_id
		hasCookie      bool   // Нужно ли вообще добавлять куку в запрос
		mockInit       mockInit
		expectedStatus int
		checkResponse  func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:        "Успешный выход",
			hasCookie:   true,
			cookieValue: uuid.New().String(),
			mockInit: func(m *ucMocks.MockAuthUseCase) {
				// Ожидаем вызов Logout в UseCase
				m.EXPECT().
					Logout(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				// Проверяем, что пришла инструкция на удаление куки
				// Мы ищем куку с тем же именем, но пустую и протухшую
				resp := rec.Result()
				cookies := resp.Cookies()
				var logoutCookie *http.Cookie
				for _, c := range cookies {
					if c.Name == "session_id" {
						logoutCookie = c
					}
				}

				assert.NotNil(t, logoutCookie)
				assert.Equal(t, "", logoutCookie.Value)
				// Проверяем, что время истечения - эпоха Unix (0)
				assert.True(t, logoutCookie.Expires.Before(time.Now()))
			},
		},
		{
			name:           "Успех: кука отсутствует",
			hasCookie:      false,
			mockInit:       func(m *ucMocks.MockAuthUseCase) {},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Успех: невалидный UUID в куке",
			hasCookie:      true,
			cookieValue:    "not-a-uuid",
			mockInit:       func(m *ucMocks.MockAuthUseCase) {},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Успех, но сессия не найдена в базе",
			hasCookie:   true,
			cookieValue: uuid.New().String(),
			mockInit: func(m *ucMocks.MockAuthUseCase) {
				// Программируем мок вернуть ошибку (например, сессия уже удалена)
				m.EXPECT().
					Logout(gomock.Any(), gomock.Any()).
					Return(domain.ErrSessionNotFound)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuthUC := ucMocks.NewMockAuthUseCase(ctrl)
			testCase.mockInit(mockAuthUC)

			mockLogger := domainMocks.NewMockLogger(ctrl)

			mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
			mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
			mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
			mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			authHandler := NewAuthHandler(mockAuthUC, mockLogger)

			req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
			if testCase.hasCookie {
				req.AddCookie(&http.Cookie{
					Name:  "session_id",
					Value: testCase.cookieValue,
				})
			}

			rec := httptest.NewRecorder()

			authHandler.Logout(rec, req)

			assert.Equal(t, testCase.expectedStatus, rec.Code)
			if testCase.checkResponse != nil {
				testCase.checkResponse(t, rec)
			}
		})
	}
}
