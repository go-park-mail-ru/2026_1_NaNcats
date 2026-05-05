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
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/userclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/userclient/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	pbUser "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/user"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUserProfileHandler_GetUserProfile(t *testing.T) {
	type mockInit func(m *mocks.MockUserClient)

	tests := []struct {
		name           string
		userID         any
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:   "Успешное получение профиля",
			userID: int64(1),
			mockInit: func(m *mocks.MockUserClient) {
				m.EXPECT().GetUserProfile(gomock.Any(), int64(1)).
					Return(&pbUser.User{Name: "Ivan", Email: "test@mail.ru", AvatarUrl: "url"}, nil, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Ошибка: пользователь не авторизован",
			userID:         nil,
			mockInit:       func(m *mocks.MockUserClient) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "Ошибка: пользователь не найден",
			userID: int64(404),
			mockInit: func(m *mocks.MockUserClient) {
				m.EXPECT().GetUserProfile(gomock.Any(), int64(404)).
					Return(nil, nil, userclient.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "Ошибка: внутренний сбой gRPC",
			userID: int64(1),
			mockInit: func(m *mocks.MockUserClient) {
				m.EXPECT().GetUserProfile(gomock.Any(), int64(1)).
					Return(nil, nil, errors.New("grpc internal"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockUserClient(ctrl)
			tt.mockInit(mockClient)

			handler := NewUserProfileHandler(mockClient, logger.NewNopLogger())

			req := httptest.NewRequest(http.MethodGet, "/api/profile", nil)
			if tt.userID != nil {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, tt.userID))
			}

			rec := httptest.NewRecorder()
			handler.GetUserProfile(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestUserProfileHandler_UpdateProfile(t *testing.T) {
	type mockInit func(m *mocks.MockUserClient)

	name := "Новое Имя"
	email := "new@mail.ru"

	tests := []struct {
		name           string
		userID         any
		idemKey        string
		body           any
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:    "Успешное обновление профиля",
			userID:  int64(1),
			idemKey: "idem-key-999",
			body: UserProfileUpdateRequest{
				Name:  &name,
				Email: &email,
			},
			mockInit: func(m *mocks.MockUserClient) {
				m.EXPECT().UpdateProfile(gomock.Any(), int64(1), &name, &email, "idem-key-999").
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Ошибка: пользователь не авторизован",
			userID:         nil,
			idemKey:        "key",
			body:           UserProfileUpdateRequest{Name: &name},
			mockInit:       func(m *mocks.MockUserClient) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Ошибка: отсутствует Idempotency-Key",
			userID:         int64(1),
			idemKey:        "",
			body:           UserProfileUpdateRequest{Name: &name},
			mockInit:       func(m *mocks.MockUserClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Ошибка: невалидный JSON",
			userID:         int64(1),
			idemKey:        "key",
			body:           "not-json",
			mockInit:       func(m *mocks.MockUserClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Ошибка: email уже занят",
			userID:  int64(1),
			idemKey: "key",
			body:    UserProfileUpdateRequest{Email: &email},
			mockInit: func(m *mocks.MockUserClient) {
				m.EXPECT().UpdateProfile(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(userclient.ErrEmailAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:    "Ошибка: пустые данные или невалидные аргументы",
			userID:  int64(1),
			idemKey: "key",
			body:    UserProfileUpdateRequest{},
			mockInit: func(m *mocks.MockUserClient) {
				m.EXPECT().UpdateProfile(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(userclient.ErrInvalidArgument)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Системная ошибка gRPC клиента",
			userID:  int64(1),
			idemKey: "key",
			body:    UserProfileUpdateRequest{Name: &name},
			mockInit: func(m *mocks.MockUserClient) {
				m.EXPECT().UpdateProfile(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("internal gRPC fail"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockUserClient(ctrl)
			tt.mockInit(mockClient)

			handler := NewUserProfileHandler(mockClient, logger.NewNopLogger())

			var buf bytes.Buffer
			if s, ok := tt.body.(string); ok {
				buf.WriteString(s)
			} else {
				_ = json.NewEncoder(&buf).Encode(tt.body)
			}

			req := httptest.NewRequest(http.MethodPatch, "/api/profile", &buf)
			if tt.userID != nil {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, tt.userID))
			}
			if tt.idemKey != "" {
				req.Header.Set("Idempotency-Key", tt.idemKey)
			}

			rec := httptest.NewRecorder()
			handler.UpdateProfile(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestUserProfileHandler_UpdateAvatar(t *testing.T) {
	type mockInit func(m *mocks.MockUserClient)

	tests := []struct {
		name           string
		userID         any
		idemKey        string
		setupForm      func() (string, io.Reader)
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:    "Успешное обновление аватара",
			userID:  int64(1),
			idemKey: "idem-123",
			setupForm: func() (string, io.Reader) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("avatar", "test.png")
				part.Write([]byte("fake-image-content"))
				writer.Close()
				return writer.FormDataContentType(), body
			},
			mockInit: func(m *mocks.MockUserClient) {
				m.EXPECT().UpdateAvatar(gomock.Any(), int64(1), []byte("fake-image-content"), "idem-123").
					Return("http://storage/new.webp", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Ошибка: пользователь не авторизован",
			userID: nil,
			setupForm: func() (string, io.Reader) {
				return "multipart/form-data", strings.NewReader("")
			},
			mockInit:       func(m *mocks.MockUserClient) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:    "Ошибка: отсутствует заголовок идемпотентности",
			userID:  int64(1),
			idemKey: "",
			setupForm: func() (string, io.Reader) {
				return "multipart/form-data", strings.NewReader("")
			},
			mockInit:       func(m *mocks.MockUserClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Ошибка: файл слишком большой",
			userID:  int64(1),
			idemKey: "idem",
			setupForm: func() (string, io.Reader) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("avatar", "huge.png")
				part.Write(make([]byte, 6*1024*1024)) // 6MB
				writer.Close()
				return writer.FormDataContentType(), body
			},
			mockInit:       func(m *mocks.MockUserClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Ошибка: отсутствует поле avatar в форме",
			userID:  int64(1),
			idemKey: "idem",
			setupForm: func() (string, io.Reader) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				writer.Close()
				return writer.FormDataContentType(), body
			},
			mockInit:       func(m *mocks.MockUserClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Ошибка: неподдерживаемый формат изображения",
			userID:  int64(1),
			idemKey: "idem",
			setupForm: func() (string, io.Reader) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("avatar", "test.txt")
				part.Write([]byte("text-content"))
				writer.Close()
				return writer.FormDataContentType(), body
			},
			mockInit: func(m *mocks.MockUserClient) {
				m.EXPECT().UpdateAvatar(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", userclient.ErrInvalidArgument)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Ошибка gRPC при загрузке",
			userID:  int64(1),
			idemKey: "idem",
			setupForm: func() (string, io.Reader) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("avatar", "test.png")
				part.Write([]byte("data"))
				writer.Close()
				return writer.FormDataContentType(), body
			},
			mockInit: func(m *mocks.MockUserClient) {
				m.EXPECT().UpdateAvatar(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", errors.New("grpc internal"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockUserClient(ctrl)
			tt.mockInit(mockClient)

			handler := NewUserProfileHandler(mockClient, logger.NewNopLogger())

			contentType, body := tt.setupForm()
			req := httptest.NewRequest(http.MethodPost, "/api/profile/avatar", body)
			req.Header.Set("Content-Type", contentType)
			if tt.userID != nil {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, tt.userID))
			}
			if tt.idemKey != "" {
				req.Header.Set("Idempotency-Key", tt.idemKey)
			}

			rec := httptest.NewRecorder()
			handler.UpdateAvatar(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestUserProfileHandler_DeleteAvatar(t *testing.T) {
	type mockInit func(m *mocks.MockUserClient)

	tests := []struct {
		name           string
		userID         any
		idemKey        string
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:    "Успешное удаление аватара",
			userID:  int64(1),
			idemKey: "idem-delete",
			mockInit: func(m *mocks.MockUserClient) {
				m.EXPECT().DeleteAvatar(gomock.Any(), int64(1), "idem-delete").
					Return("http://storage/default.webp", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Ошибка gRPC при удалении",
			userID:  int64(1),
			idemKey: "idem",
			mockInit: func(m *mocks.MockUserClient) {
				m.EXPECT().DeleteAvatar(gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", errors.New("fail"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockUserClient(ctrl)
			tt.mockInit(mockClient)

			handler := NewUserProfileHandler(mockClient, logger.NewNopLogger())

			req := httptest.NewRequest(http.MethodDelete, "/api/profile/avatar", nil)
			if tt.userID != nil {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, tt.userID))
			}
			if tt.idemKey != "" {
				req.Header.Set("Idempotency-Key", tt.idemKey)
			}

			rec := httptest.NewRecorder()
			handler.DeleteAvatar(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestUserProfileHandler_AdminUpdateRole(t *testing.T) {
	type mockInit func(m *mocks.MockUserClient)

	tests := []struct {
		name           string
		idemKey        string
		body           any
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:    "Успешное обновление роли админом",
			idemKey: "admin-key",
			body:    UpdateRoleRequest{UserID: 10, NewRole: "owner"},
			mockInit: func(m *mocks.MockUserClient) {
				m.EXPECT().UpdateRole(gomock.Any(), int64(10), "owner", "admin-key").
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Ошибка: невалидный JSON в запросе",
			idemKey:        "key",
			body:           "invalid-json",
			mockInit:       func(m *mocks.MockUserClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Ошибка gRPC сервиса: пользователь не найден",
			idemKey: "key",
			body:    UpdateRoleRequest{UserID: 5, NewRole: "support"},
			mockInit: func(m *mocks.MockUserClient) {
				m.EXPECT().UpdateRole(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(status.Error(codes.NotFound, "user not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockUserClient(ctrl)
			tt.mockInit(mockClient)

			handler := NewUserProfileHandler(mockClient, logger.NewNopLogger())

			var buf bytes.Buffer
			if s, ok := tt.body.(string); ok {
				buf.WriteString(s)
			} else {
				_ = json.NewEncoder(&buf).Encode(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/admin/users/role", &buf)
			if tt.idemKey != "" {
				req.Header.Set("Idempotency-Key", tt.idemKey)
			}

			rec := httptest.NewRecorder()
			handler.AdminUpdateRole(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}
