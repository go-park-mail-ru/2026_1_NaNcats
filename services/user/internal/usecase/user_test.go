package usecase

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/domain"
	repoMocks "github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/repository/mocks"
	ucMocks "github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/usecase/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	loggerMocks "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"
	s3Mocks "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/s3/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserUseCase_Create(t *testing.T) {
	type mockInit func(ur *repoMocks.MockUserRepository)

	ctx := context.Background()
	user := domain.User{Email: "test@mail.ru", Name: "Ivan"}
	password := "password123"
	idemKey := "idem-key"
	var expectedID int64 = 1

	tests := []struct {
		name          string
		password      string
		mockInit      mockInit
		want          int64
		expectedError error
	}{
		{
			name:     "Успешное создание",
			password: password,
			mockInit: func(ur *repoMocks.MockUserRepository) {
				// Проверяем, что в репозиторий уходит юзер с уже заполненным хэшем пароля
				ur.EXPECT().
					CreateUser(gomock.Any(), gomock.Any(), idemKey).
					DoAndReturn(func(ctx context.Context, u domain.User, key string) (int64, error) {
						if u.PasswordHash == "" {
							return 0, errors.New("password hash is empty")
						}
						return expectedID, nil
					})
			},
			want: expectedID,
		},
		{
			name:     "Ошибка: пароль слишком короткий (валидация passUtil)",
			password: "123",
			mockInit: func(ur *repoMocks.MockUserRepository) {},
			want:     0,
			// errutil.Internal так как это упало на этапе хэширования
			expectedError: errutil.Internal("failed to hash password", errors.New("password too short")),
		},
		{
			name:     "Ошибка: почта уже существует",
			password: password,
			mockInit: func(ur *repoMocks.MockUserRepository) {
				ur.EXPECT().
					CreateUser(gomock.Any(), gomock.Any(), idemKey).
					Return(int64(0), domain.ErrEmailAlreadyExists)
			},
			want:          0,
			expectedError: domain.ErrEmailAlreadyExists,
		},
		{
			name:     "Ошибка репозитория (Internal)",
			password: password,
			mockInit: func(ur *repoMocks.MockUserRepository) {
				ur.EXPECT().
					CreateUser(gomock.Any(), gomock.Any(), idemKey).
					Return(int64(0), errors.New("db error"))
			},
			want:          0,
			expectedError: errutil.Internal("failed to create user in db", errors.New("db error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ur := repoMocks.NewMockUserRepository(ctrl)
			fs := s3Mocks.NewMockFileStorage(ctrl)
			mp := ucMocks.NewMockMessagePublisher(ctrl)
			l := loggerMocks.NewNopLogger()

			tt.mockInit(ur)

			uc := NewUserUseCase(ur, fs, "default.png", mp, l)
			id, err := uc.Create(ctx, user, tt.password, idemKey)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, id)
			}
		})
	}
}

func TestUserUseCase_GetByID(t *testing.T) {
	type mockInit func(ur *repoMocks.MockUserRepository)

	ctx := context.Background()
	userID := int64(42)
	defaultAvatar := "default.png"

	tests := []struct {
		name          string
		mockInit      mockInit
		wantAvatar    string
		expectedError error
	}{
		{
			name: "Успех: аватарка есть",
			mockInit: func(ur *repoMocks.MockUserRepository) {
				ur.EXPECT().GetUserByID(gomock.Any(), userID).Return(domain.User{ID: userID, AvatarURL: "my.png"}, nil)
			},
			wantAvatar: "my.png",
		},
		{
			name: "Успех: аватарки нет, подставляем дефолтную",
			mockInit: func(ur *repoMocks.MockUserRepository) {
				ur.EXPECT().GetUserByID(gomock.Any(), userID).Return(domain.User{ID: userID, AvatarURL: ""}, nil)
			},
			wantAvatar: defaultAvatar,
		},
		{
			name: "Ошибка: пользователь не найден",
			mockInit: func(ur *repoMocks.MockUserRepository) {
				ur.EXPECT().GetUserByID(gomock.Any(), userID).Return(domain.User{}, domain.ErrUserNotFound)
			},
			expectedError: domain.ErrUserNotFound,
		},
		{
			name: "Ошибка БД",
			mockInit: func(ur *repoMocks.MockUserRepository) {
				ur.EXPECT().GetUserByID(gomock.Any(), userID).Return(domain.User{}, errors.New("db crash"))
			},
			expectedError: errutil.Internal("failed to get user from db", errors.New("db crash")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ur := repoMocks.NewMockUserRepository(ctrl)
			tt.mockInit(ur)

			uc := NewUserUseCase(ur, nil, defaultAvatar, nil, loggerMocks.NewNopLogger())
			res, err := uc.GetByID(ctx, userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantAvatar, res.AvatarURL)
			}
		})
	}
}

func TestUserUseCase_GetByEmail(t *testing.T) {
	type mockInit func(ur *repoMocks.MockUserRepository)

	ctx := context.Background()
	email := "ivan@mail.ru"
	defaultAvatar := "default.png"

	tests := []struct {
		name          string
		mockInit      mockInit
		wantAvatar    string
		expectedError error
	}{
		{
			name: "Успешное получение по email",
			mockInit: func(ur *repoMocks.MockUserRepository) {
				ur.EXPECT().GetUserByEmail(gomock.Any(), email).Return(domain.User{ID: 1, Email: email, AvatarURL: ""}, nil)
			},
			wantAvatar: defaultAvatar,
		},
		{
			name: "Пользователь не найден",
			mockInit: func(ur *repoMocks.MockUserRepository) {
				ur.EXPECT().GetUserByEmail(gomock.Any(), email).Return(domain.User{}, domain.ErrUserNotFound)
			},
			expectedError: domain.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ur := repoMocks.NewMockUserRepository(ctrl)
			tt.mockInit(ur)

			uc := NewUserUseCase(ur, nil, defaultAvatar, nil, loggerMocks.NewNopLogger())
			res, err := uc.GetByEmail(ctx, email)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantAvatar, res.AvatarURL)
			}
		})
	}
}

func TestUserUseCase_Check(t *testing.T) {
	type mockInit func(ur *repoMocks.MockUserRepository)

	ctx := context.Background()
	userID := int64(1)

	tests := []struct {
		name          string
		mockInit      mockInit
		want          bool
		expectedError error
	}{
		{
			name: "Пользователь существует",
			mockInit: func(ur *repoMocks.MockUserRepository) {
				ur.EXPECT().CheckUserByID(gomock.Any(), userID).Return(true, nil)
			},
			want: true,
		},
		{
			name: "Пользователь не найден",
			mockInit: func(ur *repoMocks.MockUserRepository) {
				ur.EXPECT().CheckUserByID(gomock.Any(), userID).Return(false, nil)
			},
			want: false,
		},
		{
			name: "Ошибка базы данных",
			mockInit: func(ur *repoMocks.MockUserRepository) {
				ur.EXPECT().CheckUserByID(gomock.Any(), userID).Return(false, errors.New("db error"))
			},
			want:          false,
			expectedError: errutil.Internal("failed to check user existence in db", errors.New("db error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ur := repoMocks.NewMockUserRepository(ctrl)
			tt.mockInit(ur)

			uc := NewUserUseCase(ur, nil, "", nil, loggerMocks.NewNopLogger())
			res, err := uc.Check(ctx, userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, res)
			}
		})
	}
}

func TestUserUseCase_UpdateProfile(t *testing.T) {
	type mockInit func(ur *repoMocks.MockUserRepository)

	ctx := context.Background()
	userID := int64(1)
	name := "NewName"
	email := "new@mail.ru"
	idemKey := "idem-update"

	tests := []struct {
		name          string
		mockInit      mockInit
		expectedError error
	}{
		{
			name: "Успешное обновление",
			mockInit: func(ur *repoMocks.MockUserRepository) {
				ur.EXPECT().UpdateProfile(gomock.Any(), userID, &name, &email).Return(nil)
			},
		},
		{
			name: "Пользователь не найден",
			mockInit: func(ur *repoMocks.MockUserRepository) {
				ur.EXPECT().UpdateProfile(gomock.Any(), userID, &name, &email).Return(domain.ErrUserNotFound)
			},
			expectedError: domain.ErrUserNotFound,
		},
		{
			name: "Email уже занят",
			mockInit: func(ur *repoMocks.MockUserRepository) {
				ur.EXPECT().UpdateProfile(gomock.Any(), userID, &name, &email).Return(domain.ErrEmailAlreadyExists)
			},
			expectedError: domain.ErrEmailAlreadyExists,
		},
		{
			name: "Внутренняя ошибка репозитория",
			mockInit: func(ur *repoMocks.MockUserRepository) {
				ur.EXPECT().UpdateProfile(gomock.Any(), userID, &name, &email).Return(errors.New("something went wrong"))
			},
			expectedError: errutil.Internal("failed to update user profile", errors.New("something went wrong")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ur := repoMocks.NewMockUserRepository(ctrl)
			tt.mockInit(ur)

			uc := NewUserUseCase(ur, nil, "", nil, loggerMocks.NewNopLogger())
			err := uc.UpdateProfile(ctx, userID, &name, &email, idemKey)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func createTestImage() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	buf := new(bytes.Buffer)
	_ = png.Encode(buf, img)
	return buf.Bytes()
}

func TestUserUseCase_UpdateAvatar(t *testing.T) {
	type mockInit func(ur *repoMocks.MockUserRepository, fs *s3Mocks.MockFileStorage)

	ctx := context.Background()
	userID := int64(1)
	idemKey := "idem-key"
	defaultAvatar := "http://s3.ru/default.webp"
	newAvatarURL := "http://s3.ru/avatars/new-uuid.webp"
	validImage := createTestImage()

	tests := []struct {
		name          string
		imageData     []byte
		mockInit      mockInit
		want          string
		expectedError string
	}{
		{
			name:      "Успешное обновление аватара (с удалением старого)",
			imageData: validImage,
			mockInit: func(ur *repoMocks.MockUserRepository, fs *s3Mocks.MockFileStorage) {
				ur.EXPECT().GetUserByID(gomock.Any(), userID).Return(domain.User{
					ID: userID, AvatarURL: "http://s3.ru/avatars/old.webp",
				}, nil)
				fs.EXPECT().UploadFile(gomock.Any(), gomock.Any(), gomock.Any(), "image/webp").Return(newAvatarURL, nil)
				ur.EXPECT().UpdateAvatarURL(gomock.Any(), userID, newAvatarURL).Return(nil)
				fs.EXPECT().DeleteFile(gomock.Any(), "http://s3.ru/avatars/old.webp").Return(nil)
			},
			want: newAvatarURL,
		},
		{
			name:      "Ошибка: невалидный формат изображения",
			imageData: []byte("not-an-image-at-all"),
			mockInit: func(ur *repoMocks.MockUserRepository, fs *s3Mocks.MockFileStorage) {
				ur.EXPECT().GetUserByID(gomock.Any(), userID).Return(domain.User{ID: userID}, nil)
			},
			expectedError: "invalid image extension",
		},
		{
			name:      "Ошибка S3: сбой при загрузке файла",
			imageData: validImage,
			mockInit: func(ur *repoMocks.MockUserRepository, fs *s3Mocks.MockFileStorage) {
				ur.EXPECT().GetUserByID(gomock.Any(), userID).Return(domain.User{ID: userID}, nil)
				fs.EXPECT().UploadFile(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", errors.New("s3 fail"))
			},
			expectedError: "failed to upload to S3 storage",
		},
		{
			name:      "Ошибка БД: очистка S3 после неудачного сохранения в БД",
			imageData: validImage,
			mockInit: func(ur *repoMocks.MockUserRepository, fs *s3Mocks.MockFileStorage) {
				ur.EXPECT().GetUserByID(gomock.Any(), userID).Return(domain.User{ID: userID}, nil)
				fs.EXPECT().UploadFile(gomock.Any(), gomock.Any(), gomock.Any(), "image/webp").Return(newAvatarURL, nil)
				ur.EXPECT().UpdateAvatarURL(gomock.Any(), userID, newAvatarURL).Return(errors.New("db error"))

				fs.EXPECT().DeleteFile(gomock.Any(), newAvatarURL).Return(nil)
			},
			expectedError: "failed to update avatar path in database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ur := repoMocks.NewMockUserRepository(ctrl)
			fs := s3Mocks.NewMockFileStorage(ctrl)
			tt.mockInit(ur, fs)

			uc := NewUserUseCase(ur, fs, defaultAvatar, nil, loggerMocks.NewNopLogger())
			res, err := uc.UpdateAvatar(ctx, userID, tt.imageData, idemKey)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, res)
			}

			time.Sleep(20 * time.Millisecond)
		})
	}
}

func TestUserUseCase_DeleteAvatar(t *testing.T) {
	type mockInit func(ur *repoMocks.MockUserRepository, fs *s3Mocks.MockFileStorage)

	ctx := context.Background()
	userID := int64(1)
	defaultAvatar := "http://s3.ru/default.webp"
	oldAvatarURL := "http://s3.ru/avatars/old.webp"

	tests := []struct {
		name          string
		mockInit      mockInit
		want          string
		expectedError error
	}{
		{
			name: "Успешное удаление",
			mockInit: func(ur *repoMocks.MockUserRepository, fs *s3Mocks.MockFileStorage) {
				ur.EXPECT().GetUserByID(gomock.Any(), userID).Return(domain.User{
					ID: userID, AvatarURL: oldAvatarURL,
				}, nil)
				ur.EXPECT().UpdateAvatarURL(gomock.Any(), userID, "").Return(nil)
				fs.EXPECT().DeleteFile(gomock.Any(), oldAvatarURL).Return(nil)
			},
			want: defaultAvatar,
		},
		{
			name: "Аватар уже дефолтный (ничего не делаем)",
			mockInit: func(ur *repoMocks.MockUserRepository, fs *s3Mocks.MockFileStorage) {
				ur.EXPECT().GetUserByID(gomock.Any(), userID).Return(domain.User{
					ID: userID, AvatarURL: defaultAvatar,
				}, nil)
			},
			want: defaultAvatar,
		},
		{
			name: "Ошибка репозитория при сбросе URL",
			mockInit: func(ur *repoMocks.MockUserRepository, fs *s3Mocks.MockFileStorage) {
				ur.EXPECT().GetUserByID(gomock.Any(), userID).Return(domain.User{
					ID: userID, AvatarURL: oldAvatarURL,
				}, nil)
				ur.EXPECT().UpdateAvatarURL(gomock.Any(), userID, "").Return(errors.New("db fail"))
			},
			expectedError: errutil.Internal("failed to reset avatar in database", errors.New("db fail")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ur := repoMocks.NewMockUserRepository(ctrl)
			fs := s3Mocks.NewMockFileStorage(ctrl)

			tt.mockInit(ur, fs)

			uc := NewUserUseCase(ur, fs, defaultAvatar, nil, loggerMocks.NewNopLogger())
			res, err := uc.DeleteAvatar(ctx, userID, "idem")

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, res)
			}

			time.Sleep(10 * time.Millisecond)
		})
	}
}

func TestUserUseCase_UpdateRole(t *testing.T) {
	type mockInit func(ur *repoMocks.MockUserRepository, mp *ucMocks.MockMessagePublisher)

	ctx := context.Background()
	userID := int64(42)
	idemKey := "role-key"

	tests := []struct {
		name          string
		newRole       string
		mockInit      mockInit
		expectedError error
	}{
		{
			name:    "Успешное изменение роли с отправкой события",
			newRole: domain.RoleCourier,
			mockInit: func(ur *repoMocks.MockUserRepository, mp *ucMocks.MockMessagePublisher) {
				ur.EXPECT().
					UpdateUserRole(gomock.Any(), userID, domain.RoleCourier, idemKey).
					Return(domain.RoleClient, true, nil)

				event := events.UserRoleChangedEvent{
					UserID:  userID,
					OldRole: domain.RoleClient,
					NewRole: domain.RoleCourier,
				}
				mp.EXPECT().
					PublishJSON(gomock.Any(), events.QueueUserEvents, event).
					Return(nil)
			},
		},
		{
			name:    "Кейс идемпотентности: повторный запрос (shouldNotify = false)",
			newRole: domain.RoleAdmin,
			mockInit: func(ur *repoMocks.MockUserRepository, mp *ucMocks.MockMessagePublisher) {
				ur.EXPECT().
					UpdateUserRole(gomock.Any(), userID, domain.RoleAdmin, idemKey).
					Return(domain.RoleAdmin, false, nil)
				// Событие не должно отправляться
			},
		},
		{
			name:    "Ошибка: невалидная роль",
			newRole: "super-god-role",
			mockInit: func(ur *repoMocks.MockUserRepository, mp *ucMocks.MockMessagePublisher) {
				// Валидация не пропустит запрос до репозитория
			},
			expectedError: domain.ErrInvalidInput,
		},
		{
			name:    "Ошибка репозитория",
			newRole: domain.RoleOwner,
			mockInit: func(ur *repoMocks.MockUserRepository, mp *ucMocks.MockMessagePublisher) {
				ur.EXPECT().
					UpdateUserRole(gomock.Any(), userID, domain.RoleOwner, idemKey).
					Return("", false, errors.New("db error"))
			},
			expectedError: errors.New("db error"),
		},
		{
			name:    "Частичный успех: БД обновилась, но RabbitMQ упал (ошибка не возвращается)",
			newRole: domain.RoleSupport,
			mockInit: func(ur *repoMocks.MockUserRepository, mp *ucMocks.MockMessagePublisher) {
				ur.EXPECT().
					UpdateUserRole(gomock.Any(), userID, domain.RoleSupport, idemKey).
					Return(domain.RoleClient, true, nil)

				mp.EXPECT().
					PublishJSON(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("rabbit unreachable"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ur := repoMocks.NewMockUserRepository(ctrl)
			mp := ucMocks.NewMockMessagePublisher(ctrl)
			l := loggerMocks.NewNopLogger()

			tt.mockInit(ur, mp)

			uc := NewUserUseCase(ur, nil, "", mp, l)
			err := uc.UpdateRole(ctx, userID, tt.newRole, idemKey)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
