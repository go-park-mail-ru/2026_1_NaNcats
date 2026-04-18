package usecase

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	repoMocks "github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/domain"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserUseCase_Create(t *testing.T) {
	ctx := context.Background()
	user := domain.User{Email: "test@mail.ru"}

	tests := []struct {
		name     string
		mockInit func(ur *repoMocks.MockUserRepository)
		wantID   int
		wantErr  error
	}{
		{
			name: "Успешное создание пользователя",
			mockInit: func(ur *repoMocks.MockUserRepository) {
				ur.EXPECT().CreateUser(gomock.Any(), user).Return(1, nil)
			},
			wantID:  1,
			wantErr: nil,
		},
		{
			name: "Ошибка при создании: Email уже существует",
			mockInit: func(ur *repoMocks.MockUserRepository) {
				ur.EXPECT().CreateUser(gomock.Any(), user).Return(0, domain.ErrEmailAlreadyExists)
			},
			wantID:  0,
			wantErr: domain.ErrEmailAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ur := repoMocks.NewMockUserRepository(ctrl)
			uc := NewUserUseCase(ur, nil, "")

			tt.mockInit(ur)
			id, err := uc.Create(ctx, user)

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.wantID, id)
		})
	}
}

func TestUserUseCase_GetByID(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		userID   int
		mockInit func(ur *repoMocks.MockUserRepository)
		wantErr  error
	}{
		{
			name:   "Успешное получение по ID",
			userID: 1,
			mockInit: func(ur *repoMocks.MockUserRepository) {
				ur.EXPECT().GetUserByID(gomock.Any(), 1).Return(domain.User{ID: 1}, nil)
			},
			wantErr: nil,
		},
		{
			name:   "Пользователь не найден",
			userID: 404,
			mockInit: func(ur *repoMocks.MockUserRepository) {
				ur.EXPECT().GetUserByID(gomock.Any(), 404).Return(domain.User{}, domain.ErrUserNotFound)
			},
			wantErr: domain.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ur := repoMocks.NewMockUserRepository(ctrl)
			uc := NewUserUseCase(ur, nil, "")

			tt.mockInit(ur)
			_, err := uc.GetByID(ctx, tt.userID)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestUserUseCase_Check(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		userID   int
		mockInit func(ur *repoMocks.MockUserRepository)
		want     bool
		wantErr  bool
	}{
		{
			name:   "Пользователь существует",
			userID: 1,
			mockInit: func(ur *repoMocks.MockUserRepository) {
				ur.EXPECT().CheckUserByID(gomock.Any(), 1).Return(true, nil)
			},
			want:    true,
			wantErr: false,
		},
		{
			name:   "Ошибка базы данных",
			userID: 1,
			mockInit: func(ur *repoMocks.MockUserRepository) {
				ur.EXPECT().CheckUserByID(gomock.Any(), 1).Return(false, errors.New("db fail"))
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ur := repoMocks.NewMockUserRepository(ctrl)
			uc := NewUserUseCase(ur, nil, "")

			tt.mockInit(ur)
			res, err := uc.Check(ctx, tt.userID)
			assert.Equal(t, tt.want, res)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestUserUseCase_UpdateProfile(t *testing.T) {
	ctx := context.Background()
	name := "NewName"

	tests := []struct {
		name     string
		mockInit func(ur *repoMocks.MockUserRepository)
		wantErr  error
	}{
		{
			name: "Успешное обновление профиля",
			mockInit: func(ur *repoMocks.MockUserRepository) {
				ur.EXPECT().UpdateProfile(gomock.Any(), 1, &name, nil).Return(nil)
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ur := repoMocks.NewMockUserRepository(ctrl)
			uc := NewUserUseCase(ur, nil, "")

			tt.mockInit(ur)
			err := uc.UpdateProfile(ctx, 1, &name, nil)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestUserUseCase_UpdateAvatar(t *testing.T) {
	ctx := context.Background()
	userID := 1
	validReader := strings.NewReader("fake-image-data")

	tests := []struct {
		name     string
		reader   io.Reader
		mockInit func(ur *repoMocks.MockUserRepository, fs *repoMocks.MockFileStorage)
		wantErr  error
	}{
		{
			name:   "Ошибка: пользователь не найден",
			reader: validReader,
			mockInit: func(ur *repoMocks.MockUserRepository, fs *repoMocks.MockFileStorage) {
				ur.EXPECT().GetUserByID(gomock.Any(), userID).Return(domain.User{}, domain.ErrUserNotFound)
			},
			wantErr: domain.ErrUserNotFound,
		},
		{
			name:   "Ошибка: неверный формат изображения",
			reader: strings.NewReader("not-an-image"),
			mockInit: func(ur *repoMocks.MockUserRepository, fs *repoMocks.MockFileStorage) {
				ur.EXPECT().GetUserByID(gomock.Any(), userID).Return(domain.User{ID: userID}, nil)
			},
			wantErr: domain.ErrInvalidImageExt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ur := repoMocks.NewMockUserRepository(ctrl)
			fs := repoMocks.NewMockFileStorage(ctrl)
			uc := NewUserUseCase(ur, fs, "")

			tt.mockInit(ur, fs)
			_, err := uc.UpdateAvatar(ctx, userID, tt.reader)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}

func TestUserUseCase_DeleteAvatar(t *testing.T) {
	ctx := context.Background()
	userID := 1

	tests := []struct {
		name     string
		mockInit func(ur *repoMocks.MockUserRepository, fs *repoMocks.MockFileStorage)
		wantErr  error
	}{
		{
			name: "Успешное удаление аватара",
			mockInit: func(ur *repoMocks.MockUserRepository, fs *repoMocks.MockFileStorage) {
				ur.EXPECT().GetUserByID(gomock.Any(), userID).Return(domain.User{AvatarURL: "old.webp"}, nil)
				ur.EXPECT().UpdateAvatarURL(gomock.Any(), userID, "").Return(nil)
				fs.EXPECT().DeleteFile(gomock.Any(), "old.webp").Return(nil).AnyTimes()
			},
			wantErr: nil,
		},
		{
			name: "Аватар уже отсутствует",
			mockInit: func(ur *repoMocks.MockUserRepository, fs *repoMocks.MockFileStorage) {
				ur.EXPECT().GetUserByID(gomock.Any(), userID).Return(domain.User{AvatarURL: ""}, nil)
			},
			wantErr: nil,
		},
		{
			name: "Ошибка БД при обновлении URL",
			mockInit: func(ur *repoMocks.MockUserRepository, fs *repoMocks.MockFileStorage) {
				ur.EXPECT().GetUserByID(gomock.Any(), userID).Return(domain.User{AvatarURL: "old.webp"}, nil)
				ur.EXPECT().UpdateAvatarURL(gomock.Any(), userID, "").Return(errors.New("db fail"))
			},
			wantErr: errors.New("db fail"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ur := repoMocks.NewMockUserRepository(ctrl)
			fs := repoMocks.NewMockFileStorage(ctrl)
			uc := NewUserUseCase(ur, fs, "")

			tt.mockInit(ur, fs)
			_, err := uc.DeleteAvatar(ctx, userID)

			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			time.Sleep(time.Millisecond * 5)
		})
	}
}
