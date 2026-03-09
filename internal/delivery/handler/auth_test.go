package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthHandler_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthUC := mocks.NewMockAuthUseCase(ctrl)

	authHandler := NewAuthHandler(mockAuthUC)

	t.Run("Успешная регистрация", func(t *testing.T) {
		reqBody := RegisterRequest{
			Name:     "Ivan",
			Email:    "test@mail.ru",
			Password: "password123",
		}

		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(reqBody)
		assert.NoError(t, err)

		// Создаем виртуальный запрос и рекордер (куда запишется ответ)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", &buf)
		rec := httptest.NewRecorder()

		mockUser := domain.User{ID: uuid.New(), Name: "Ivan", Email: "test@mail.ru"}
		mockSess := domain.Session{ID: uuid.New(), ExpiresAt: time.Now().Add(time.Hour)}

		mockAuthUC.EXPECT().
			Register(gomock.Any(), gomock.Any()).
			Return(mockUser, mockSess, nil)

		authHandler.Register(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Contains(t, rec.Header().Get("Set-Cookie"), "session_id")

		var resp RegisterResponse
		err = json.NewDecoder(rec.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, mockUser.Name, resp.Name)
	})

	t.Run("Неверный JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader("invalid-json"))
		rec := httptest.NewRecorder()

		authHandler.Register(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestAuthHandler_GetMe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthUC := mocks.NewMockAuthUseCase(ctrl)

	authHandler := NewAuthHandler(mockAuthUC)

	t.Run("Успешный запуск", func(t *testing.T) {
		userID := uuid.New()

		req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()

		mockAuthUC.EXPECT().
			GetProfile(gomock.Any(), userID).
			Return(domain.User{ID: userID, Name: "Ivan"}, nil)

		authHandler.GetMe(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Ivan")
	})

	t.Run("No User in Context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		rec := httptest.NewRecorder()

		authHandler.GetMe(rec, req)

		// код в GetMe при !ok возвращает StatusInternalServerError
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthUC := mocks.NewMockAuthUseCase(ctrl)

	authHandler := NewAuthHandler(mockAuthUC)

	t.Run("Успешный вход", func(t *testing.T) {
		reqBody := LoginRequest{Login: "test@gmail.com", Password: "aboba"}

		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(reqBody)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", &buf)
		rec := httptest.NewRecorder()

		mockUser := domain.User{ID: uuid.New(), Name: "Ivan", Email: "test@mail.ru"}
		mockSess := domain.Session{ID: uuid.New(), ExpiresAt: time.Now().Add(time.Hour)}

		mockAuthUC.EXPECT().Login(gomock.Any(), gomock.Any()).Return(mockUser, mockSess, nil)

		authHandler.Login(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get("Set-Cookie"), "session_id")

		var resp LoginResponse
		err = json.NewDecoder(rec.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, mockUser.Name, resp.Name)
	})

	t.Run("Неверный JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader("invalid JSON"))
		rec := httptest.NewRecorder()

		authHandler.Login(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Неверный пароль", func(t *testing.T) {
		reqBody := LoginRequest{Login: "test@mail.ru", Password: "wrong"}
		var buf bytes.Buffer
		json.NewEncoder(&buf).Encode(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", &buf)
		rec := httptest.NewRecorder()

		// Программируем UseCase вернуть ошибку
		mockAuthUC.EXPECT().
			Login(gomock.Any(), gomock.Any()).
			Return(domain.User{}, domain.Session{}, bcrypt.ErrMismatchedHashAndPassword)

		authHandler.Login(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthUC := mocks.NewMockAuthUseCase(ctrl)
	authHandler := NewAuthHandler(mockAuthUC)

	t.Run("Успешный выход", func(t *testing.T) {
		sessionID := uuid.New()
		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)

		// Добавляем куку в запрос
		req.AddCookie(&http.Cookie{
			Name:  "session_id",
			Value: sessionID.String(),
		})

		rec := httptest.NewRecorder()

		// Ожидаем вызов Logout в UseCase
		mockAuthUC.EXPECT().
			Logout(gomock.Any(), sessionID).
			Return(nil)

		authHandler.Logout(rec, req)

		// 1. Проверяем статус
		assert.Equal(t, http.StatusOK, rec.Code)

		// 2. Проверяем, что пришла инструкция на удаление куки
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
	})

	t.Run("Ошибка: кука отсутствует", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		rec := httptest.NewRecorder()

		authHandler.Logout(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Ошибка: невалидный UUID в куке", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_id",
			Value: "not-a-uuid",
		})
		rec := httptest.NewRecorder()

		authHandler.Logout(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Ошибка: сессия не найдена в базе", func(t *testing.T) {
		sessionID := uuid.New()
		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_id",
			Value: sessionID.String(),
		})
		rec := httptest.NewRecorder()

		// Программируем мок вернуть ошибку (например, сессия уже удалена)
		mockAuthUC.EXPECT().
			Logout(gomock.Any(), sessionID).
			Return(domain.ErrSessionNotFound)

		authHandler.Logout(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
