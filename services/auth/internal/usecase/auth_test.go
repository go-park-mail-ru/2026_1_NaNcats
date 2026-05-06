package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/auth/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/auth/internal/usecase/mocks"
	passUtil "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/password"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
)

type statusCoder interface {
	GRPCStatus() codes.Code
}

func TestAuthUseCase_IssueSession(t *testing.T) {
	type mockInit func(userMock *mocks.MockUserClient, sessMock *mocks.MockSessionUseCase)

	sessID := uuid.New()
	tests := []struct {
		name        string
		userID      int64
		role        string
		userAgent   string
		mockInit    mockInit
		expectedID  uuid.UUID
		expectedErr bool
	}{
		{
			name:      "Успешное создание сессии",
			userID:    1,
			role:      "user",
			userAgent: "Mozilla",
			mockInit: func(userMock *mocks.MockUserClient, sessMock *mocks.MockSessionUseCase) {
				sessMock.EXPECT().Create(gomock.Any(), int64(1), "user", "Mozilla").
					Return(domain.Session{ID: sessID, UserID: 1}, nil)
			},
			expectedID:  sessID,
			expectedErr: false,
		},
		{
			name:      "Ошибка создания в SessionUseCase",
			userID:    1,
			role:      "user",
			userAgent: "Mozilla",
			mockInit: func(userMock *mocks.MockUserClient, sessMock *mocks.MockSessionUseCase) {
				sessMock.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(domain.Session{}, errors.New("redis timeout"))
			},
			expectedID:  uuid.Nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userMock := mocks.NewMockUserClient(ctrl)
			sessMock := mocks.NewMockSessionUseCase(ctrl)
			tt.mockInit(userMock, sessMock)

			uc := NewAuthUseCase(userMock, sessMock)
			sess, err := uc.IssueSession(context.Background(), tt.userID, tt.role, tt.userAgent)

			if tt.expectedErr {
				assert.Error(t, err)

				// Проверяем, что ошибка обернута в errutil
				domainErr, ok := err.(statusCoder)
				assert.True(t, ok)
				assert.Equal(t, codes.Internal, domainErr.GRPCStatus())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, sess.ID)
			}
		})
	}
}

func TestAuthUseCase_Login(t *testing.T) {
	type mockInit func(userMock *mocks.MockUserClient, sessMock *mocks.MockSessionUseCase)

	// Генерируем реальный валидный argon2id хэш для пароля "secret"
	validHash, err := passUtil.HashPassword("secret", nil)
	require.NoError(t, err, "Не удалось сгенерировать хэш для теста")

	sessID := uuid.New()

	tests := []struct {
		name        string
		email       string
		password    string
		userAgent   string
		mockInit    mockInit
		expectedErr bool
		errCode     codes.Code
	}{
		{
			name:      "Успешный логин",
			email:     "test@mail.ru",
			password:  "secret",
			userAgent: "Chrome",
			mockInit: func(userMock *mocks.MockUserClient, sessMock *mocks.MockSessionUseCase) {
				userMock.EXPECT().GetUserByEmail(gomock.Any(), "test@mail.ru").
					Return(domain.User{ID: 10, Email: "test@mail.ru", PasswordHash: validHash, Role: "user"}, nil)

				sessMock.EXPECT().Create(gomock.Any(), int64(10), "user", "Chrome").
					Return(domain.Session{ID: sessID, UserID: 10}, nil)
			},
			expectedErr: false,
		},
		{
			name:      "Ошибка: неверный пароль",
			email:     "test@mail.ru",
			password:  "wrong_password",
			userAgent: "Chrome",
			mockInit: func(userMock *mocks.MockUserClient, sessMock *mocks.MockSessionUseCase) {
				userMock.EXPECT().GetUserByEmail(gomock.Any(), "test@mail.ru").
					Return(domain.User{ID: 10, PasswordHash: validHash}, nil)
				// Create не должен вызываться
			},
			expectedErr: true,
			errCode:     codes.Unauthenticated,
		},
		{
			name:      "Ошибка: пользователь не найден",
			email:     "notfound@mail.ru",
			password:  "secret",
			userAgent: "Chrome",
			mockInit: func(userMock *mocks.MockUserClient, sessMock *mocks.MockSessionUseCase) {
				userMock.EXPECT().GetUserByEmail(gomock.Any(), "notfound@mail.ru").
					Return(domain.User{}, errors.New("not found in db"))
			},
			expectedErr: true,
			errCode:     codes.Unauthenticated,
		},
		{
			name:      "Ошибка создания сессии после успешной валидации",
			email:     "test@mail.ru",
			password:  "secret",
			userAgent: "Chrome",
			mockInit: func(userMock *mocks.MockUserClient, sessMock *mocks.MockSessionUseCase) {
				userMock.EXPECT().GetUserByEmail(gomock.Any(), "test@mail.ru").
					Return(domain.User{ID: 10, PasswordHash: validHash, Role: "user"}, nil)

				sessMock.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(domain.Session{}, errors.New("redis fail"))
			},
			expectedErr: true,
			errCode:     codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userMock := mocks.NewMockUserClient(ctrl)
			sessMock := mocks.NewMockSessionUseCase(ctrl)
			tt.mockInit(userMock, sessMock)

			uc := NewAuthUseCase(userMock, sessMock)
			sess, err := uc.Login(context.Background(), tt.email, tt.password, tt.userAgent)

			if tt.expectedErr {
				assert.Error(t, err)
				st, ok := err.(statusCoder)
				assert.True(t, ok)
				assert.Equal(t, tt.errCode, st.GRPCStatus())
				assert.Empty(t, sess.ID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, sessID, sess.ID)
			}
		})
	}
}

func TestAuthUseCase_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userMock := mocks.NewMockUserClient(ctrl)
	sessMock := mocks.NewMockSessionUseCase(ctrl)
	uc := NewAuthUseCase(userMock, sessMock)

	sessID := uuid.New()

	t.Run("Успешный логаут", func(t *testing.T) {
		sessMock.EXPECT().Destroy(gomock.Any(), sessID).Return(nil)
		err := uc.Logout(context.Background(), sessID)
		assert.NoError(t, err)
	})

	t.Run("Ошибка удаления сессии", func(t *testing.T) {
		sessMock.EXPECT().Destroy(gomock.Any(), sessID).Return(errors.New("db error"))
		err := uc.Logout(context.Background(), sessID)
		assert.Error(t, err)
	})
}

func TestAuthUseCase_CheckUserSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userMock := mocks.NewMockUserClient(ctrl)
	sessMock := mocks.NewMockSessionUseCase(ctrl)
	uc := NewAuthUseCase(userMock, sessMock)

	sessID := uuid.New()

	t.Run("Сессия валидна", func(t *testing.T) {
		sessMock.EXPECT().Check(gomock.Any(), sessID).
			Return(domain.Session{UserID: 42, Role: "admin"}, nil)

		userID, role, err := uc.CheckUserSession(context.Background(), sessID)
		assert.NoError(t, err)
		assert.Equal(t, int64(42), userID)
		assert.Equal(t, "admin", role)
	})

	t.Run("Сессия протухла", func(t *testing.T) {
		sessMock.EXPECT().Check(gomock.Any(), sessID).
			Return(domain.Session{}, domain.ErrSessionExpired)

		userID, role, err := uc.CheckUserSession(context.Background(), sessID)
		assert.ErrorIs(t, err, domain.ErrSessionExpired)
		assert.Equal(t, int64(0), userID)
		assert.Empty(t, role)
	})
}

func TestAuthUseCase_CSRF(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userMock := mocks.NewMockUserClient(ctrl)
	sessMock := mocks.NewMockSessionUseCase(ctrl)
	uc := NewAuthUseCase(userMock, sessMock)

	sessID := uuid.New()
	token := "csrf-token-123"

	t.Run("SetCSRFForUser", func(t *testing.T) {
		sessMock.EXPECT().SetCSRF(gomock.Any(), sessID).Return(token, nil)
		res, err := uc.SetCSRFForUser(context.Background(), sessID)
		assert.NoError(t, err)
		assert.Equal(t, token, res)
	})

	t.Run("GetCSRFBySessionID", func(t *testing.T) {
		sessMock.EXPECT().GetCSRF(gomock.Any(), sessID).Return(token, nil)
		res, err := uc.GetCSRFBySessionID(context.Background(), sessID)
		assert.NoError(t, err)
		assert.Equal(t, token, res)
	})
}
