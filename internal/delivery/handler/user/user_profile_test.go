package user

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	domainMocks "github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain/mocks"
	ucMocks "github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserProfileHandler_GetUserProfile(t *testing.T) {
	type mockInit func(up *ucMocks.MockUserProfileUseCase)

	tests := []struct {
		name           string
		userID         any
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:   "Успешное получение профиля",
			userID: 1,
			mockInit: func(up *ucMocks.MockUserProfileUseCase) {
				up.EXPECT().
					GetUserProfile(gomock.Any(), 1).
					Return(domain.User{Name: "Ivan", Email: "test@mail.ru", AvatarURL: "avatar.png"}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Пользователь не найден",
			userID: 1,
			mockInit: func(up *ucMocks.MockUserProfileUseCase) {
				up.EXPECT().
					GetUserProfile(gomock.Any(), 1).
					Return(domain.User{}, domain.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Ошибка авторизации (отсутствует ID)",
			userID:         nil,
			mockInit:       func(up *ucMocks.MockUserProfileUseCase) {},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			upuc := ucMocks.NewMockUserProfileUseCase(ctrl)
			uuc := ucMocks.NewMockUserUseCase(ctrl)
			suc := ucMocks.NewMockSessionUseCase(ctrl)
			h := NewUserProfileHandler(upuc, uuc, suc, domainMocks.NewNopLogger())

			req := httptest.NewRequest(http.MethodGet, "/profile", nil)
			if tt.userID != nil {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			tt.mockInit(upuc)

			h.GetUserProfile(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestUserProfileHandler_UpdateProfile(t *testing.T) {
	type mockInit func(u *ucMocks.MockUserUseCase)

	name := "New Name"
	email := "new@mail.ru"

	tests := []struct {
		name           string
		userID         any
		body           any
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:   "Успешное обновление",
			userID: 1,
			body:   UserProfileUpdateRequest{Name: &name, Email: &email},
			mockInit: func(u *ucMocks.MockUserUseCase) {
				u.EXPECT().UpdateProfile(gomock.Any(), 1, &name, &email).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Конфликт email",
			userID: 1,
			body:   UserProfileUpdateRequest{Email: &email},
			mockInit: func(u *ucMocks.MockUserUseCase) {
				u.EXPECT().UpdateProfile(gomock.Any(), 1, nil, &email).Return(domain.ErrEmailAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:   "Пустой запрос на обновление",
			userID: 1,
			body:   UserProfileUpdateRequest{},
			mockInit: func(u *ucMocks.MockUserUseCase) {
				u.EXPECT().UpdateProfile(gomock.Any(), 1, nil, nil).Return(domain.ErrEmptyDBQuery)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			upuc := ucMocks.NewMockUserProfileUseCase(ctrl)
			uuc := ucMocks.NewMockUserUseCase(ctrl)
			suc := ucMocks.NewMockSessionUseCase(ctrl)
			h := NewUserProfileHandler(upuc, uuc, suc, domainMocks.NewNopLogger())

			b, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPatch, "/profile", bytes.NewBuffer(b))
			if tt.userID != nil {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			tt.mockInit(uuc)

			h.UpdateProfile(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestUserProfileHandler_UpdateAvatar(t *testing.T) {
	type mockInit func(u *ucMocks.MockUserUseCase)

	tests := []struct {
		name           string
		userID         any
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:   "Успешное обновление аватара",
			userID: 1,
			mockInit: func(u *ucMocks.MockUserUseCase) {
				u.EXPECT().UpdateAvatar(gomock.Any(), 1, gomock.Any()).Return("new_url.png", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Неверный формат изображения",
			userID: 1,
			mockInit: func(u *ucMocks.MockUserUseCase) {
				u.EXPECT().UpdateAvatar(gomock.Any(), 1, gomock.Any()).Return("", domain.ErrInvalidImageExt)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			upuc := ucMocks.NewMockUserProfileUseCase(ctrl)
			uuc := ucMocks.NewMockUserUseCase(ctrl)
			suc := ucMocks.NewMockSessionUseCase(ctrl)
			h := NewUserProfileHandler(upuc, uuc, suc, domainMocks.NewNopLogger())

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, _ := writer.CreateFormFile("avatar", "test.png")
			_, _ = io.WriteString(part, "fake image content")
			writer.Close()

			req := httptest.NewRequest(http.MethodPost, "/profile/avatar", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			if tt.userID != nil {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			tt.mockInit(uuc)

			h.UpdateAvatar(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestUserProfileHandler_DeleteAvatar(t *testing.T) {
	type mockInit func(u *ucMocks.MockUserUseCase)

	defaultURL := os.Getenv("DEFAULT_AVATAR_URL")

	tests := []struct {
		name           string
		userID         any
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:   "Успешное удаление аватара",
			userID: 1,
			mockInit: func(u *ucMocks.MockUserUseCase) {
				u.EXPECT().DeleteAvatar(gomock.Any(), 1).Return(defaultURL, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Ошибка сервера при удалении",
			userID: 1,
			mockInit: func(u *ucMocks.MockUserUseCase) {
				u.EXPECT().DeleteAvatar(gomock.Any(), 1).Return("", errors.New("fail"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			upuc := ucMocks.NewMockUserProfileUseCase(ctrl)
			uuc := ucMocks.NewMockUserUseCase(ctrl)
			suc := ucMocks.NewMockSessionUseCase(ctrl)
			h := NewUserProfileHandler(upuc, uuc, suc, domainMocks.NewNopLogger())

			req := httptest.NewRequest(http.MethodDelete, "/profile/avatar", nil)
			if tt.userID != nil {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			tt.mockInit(uuc)

			h.DeleteAvatar(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
