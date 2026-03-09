package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
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
