package userclient

import (
	"context"
	"errors"
	"testing"

	pbUser "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/user"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/user/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:generate mockgen -destination=../../../../shared/proto/user/mocks/user_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/user UserServiceClient

func TestUserClient_CreateUser(t *testing.T) {
	type mockInit func(m *mocks.MockUserServiceClient)

	tests := []struct {
		name     string
		mockInit mockInit
		want     int64
		wantErr  error
	}{
		{
			name: "Успешное создание пользователя",
			mockInit: func(m *mocks.MockUserServiceClient) {
				m.EXPECT().CreateUser(gomock.Any(), &pbUser.CreateUserRequest{
					Name:           "Ivan",
					Email:          "test@mail.ru",
					Password:       "pass1234",
					IdempotencyKey: "key1",
				}).Return(&pbUser.CreateUserResponse{UserId: 1}, nil)
			},
			want:    1,
			wantErr: nil,
		},
		{
			name: "Ошибка: почта уже занята",
			mockInit: func(m *mocks.MockUserServiceClient) {
				m.EXPECT().CreateUser(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.AlreadyExists, "exists"))
			},
			want:    0,
			wantErr: ErrEmailAlreadyExists,
		},
		{
			name: "Системная ошибка gRPC",
			mockInit: func(m *mocks.MockUserServiceClient) {
				m.EXPECT().CreateUser(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("rpc error"))
			},
			want:    0,
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockUserServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewUserClient(mockSvc)
			res, err := client.CreateUser(context.Background(), "Ivan", "test@mail.ru", "pass1234", "key1")

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, res)
		})
	}
}

func TestUserClient_GetByID(t *testing.T) {
	type mockInit func(m *mocks.MockUserServiceClient)

	tests := []struct {
		name     string
		userID   int64
		mockInit mockInit
		wantErr  error
	}{
		{
			name:   "Успешное получение пользователя",
			userID: 1,
			mockInit: func(m *mocks.MockUserServiceClient) {
				m.EXPECT().GetByID(gomock.Any(), &pbUser.GetUserByIDRequest{UserId: 1}).
					Return(&pbUser.GetUserResponse{User: &pbUser.User{Id: 1, Name: "Ivan"}}, nil)
			},
			wantErr: nil,
		},
		{
			name:   "Ошибка: пользователь не найден",
			userID: 404,
			mockInit: func(m *mocks.MockUserServiceClient) {
				m.EXPECT().GetByID(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			wantErr: ErrUserNotFound,
		},
		{
			name:   "Ошибка: системный сбой",
			userID: 1,
			mockInit: func(m *mocks.MockUserServiceClient) {
				m.EXPECT().GetByID(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.Internal, "crash"))
			},
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockUserServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewUserClient(mockSvc)
			res, err := client.GetByID(context.Background(), tt.userID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, tt.userID, res.Id)
			}
		})
	}
}

func TestUserClient_GetUserProfile(t *testing.T) {
	type mockInit func(m *mocks.MockUserServiceClient)

	tests := []struct {
		name     string
		userID   int64
		mockInit mockInit
		wantErr  error
	}{
		{
			name:   "Успешное получение профиля",
			userID: 1,
			mockInit: func(m *mocks.MockUserServiceClient) {
				m.EXPECT().GetUserProfile(gomock.Any(), &pbUser.GetUserProfileRequest{UserId: 1}).
					Return(&pbUser.GetUserProfileResponse{
						User:    &pbUser.User{Id: 1},
						Profile: &pbUser.ClientProfile{AccountId: 1, BonusBalance: 100},
					}, nil)
			},
			wantErr: nil,
		},
		{
			name:   "Ошибка: профиль не найден",
			userID: 404,
			mockInit: func(m *mocks.MockUserServiceClient) {
				m.EXPECT().GetUserProfile(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			wantErr: ErrUserNotFound,
		},
		{
			name:   "Ошибка: ошибка gRPC",
			userID: 1,
			mockInit: func(m *mocks.MockUserServiceClient) {
				m.EXPECT().GetUserProfile(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("fail"))
			},
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockUserServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewUserClient(mockSvc)
			u, p, err := client.GetUserProfile(context.Background(), tt.userID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, u)
				assert.Nil(t, p)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, u)
				assert.NotNil(t, p)
				assert.Equal(t, tt.userID, u.Id)
			}
		})
	}
}

func TestUserClient_UpdateProfile(t *testing.T) {
	type mockInit func(m *mocks.MockUserServiceClient)

	name := "New Name"
	email := "new@mail.ru"

	tests := []struct {
		name     string
		mockInit mockInit
		wantErr  error
	}{
		{
			name: "Успешное обновление профиля",
			mockInit: func(m *mocks.MockUserServiceClient) {
				m.EXPECT().UpdateProfile(gomock.Any(), &pbUser.UpdateProfileRequest{
					UserId:         1,
					Name:           &name,
					Email:          &email,
					IdempotencyKey: "key-upd",
				}).Return(nil, nil)
			},
			wantErr: nil,
		},
		{
			name: "Ошибка: email уже занят",
			mockInit: func(m *mocks.MockUserServiceClient) {
				m.EXPECT().UpdateProfile(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.AlreadyExists, "email exists"))
			},
			wantErr: ErrEmailAlreadyExists,
		},
		{
			name: "Ошибка: невалидные входные данные",
			mockInit: func(m *mocks.MockUserServiceClient) {
				m.EXPECT().UpdateProfile(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.InvalidArgument, "invalid"))
			},
			wantErr: ErrInvalidArgument,
		},
		{
			name: "Ошибка: системный сбой gRPC",
			mockInit: func(m *mocks.MockUserServiceClient) {
				m.EXPECT().UpdateProfile(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("rpc fail"))
			},
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockUserServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewUserClient(mockSvc)
			err := client.UpdateProfile(context.Background(), 1, &name, &email, "key-upd")

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestUserClient_UpdateAvatar(t *testing.T) {
	type mockInit func(m *mocks.MockUserServiceClient)

	fileData := []byte("fake-image-content")

	tests := []struct {
		name     string
		mockInit mockInit
		wantUrl  string
		wantErr  error
	}{
		{
			name: "Успешное обновление аватара",
			mockInit: func(m *mocks.MockUserServiceClient) {
				m.EXPECT().UpdateAvatar(gomock.Any(), &pbUser.UpdateAvatarRequest{
					UserId:         1,
					ImageData:      fileData,
					IdempotencyKey: "avatar-key",
				}).Return(&pbUser.UpdateAvatarResponse{AvatarUrl: "http://s3/new.webp"}, nil)
			},
			wantUrl: "http://s3/new.webp",
			wantErr: nil,
		},
		{
			name: "Ошибка: неверный формат изображения",
			mockInit: func(m *mocks.MockUserServiceClient) {
				m.EXPECT().UpdateAvatar(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.InvalidArgument, "bad image"))
			},
			wantUrl: "",
			wantErr: ErrInvalidArgument,
		},
		{
			name: "Ошибка: внутренняя ошибка сервера",
			mockInit: func(m *mocks.MockUserServiceClient) {
				m.EXPECT().UpdateAvatar(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("internal"))
			},
			wantUrl: "",
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockUserServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewUserClient(mockSvc)
			url, err := client.UpdateAvatar(context.Background(), 1, fileData, "avatar-key")

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.wantUrl, url)
		})
	}
}

func TestUserClient_DeleteAvatar(t *testing.T) {
	type mockInit func(m *mocks.MockUserServiceClient)

	tests := []struct {
		name     string
		mockInit mockInit
		want     string
		wantErr  error
	}{
		{
			name: "Успешное удаление аватара",
			mockInit: func(m *mocks.MockUserServiceClient) {
				m.EXPECT().DeleteAvatar(gomock.Any(), &pbUser.DeleteAvatarRequest{
					UserId:         1,
					IdempotencyKey: "key-del",
				}).Return(&pbUser.DeleteAvatarResponse{DefaultAvatarUrl: "default.webp"}, nil)
			},
			want:    "default.webp",
			wantErr: nil,
		},
		{
			name: "Ошибка gRPC при удалении",
			mockInit: func(m *mocks.MockUserServiceClient) {
				m.EXPECT().DeleteAvatar(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("rpc fail"))
			},
			want:    "",
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockUserServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewUserClient(mockSvc)
			res, err := client.DeleteAvatar(context.Background(), 1, "key-del")

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, res)
		})
	}
}

func TestUserClient_UpdateRole(t *testing.T) {
	type mockInit func(m *mocks.MockUserServiceClient)

	tests := []struct {
		name     string
		mockInit mockInit
		wantErr  error
	}{
		{
			name: "Успешное обновление роли",
			mockInit: func(m *mocks.MockUserServiceClient) {
				m.EXPECT().UpdateUserRole(gomock.Any(), &pbUser.UpdateUserRoleRequest{
					UserId:         1,
					NewRole:        "admin",
					IdempotencyKey: "key-role",
				}).Return(nil, nil)
			},
			wantErr: nil,
		},
		{
			name: "Ошибка: невалидные аргументы",
			mockInit: func(m *mocks.MockUserServiceClient) {
				m.EXPECT().UpdateUserRole(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.InvalidArgument, "bad role"))
			},
			wantErr: ErrInvalidArgument,
		},
		{
			name: "Ошибка: пользователь не найден",
			mockInit: func(m *mocks.MockUserServiceClient) {
				m.EXPECT().UpdateUserRole(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			wantErr: ErrUserNotFound,
		},
		{
			name: "Системная ошибка gRPC",
			mockInit: func(m *mocks.MockUserServiceClient) {
				m.EXPECT().UpdateUserRole(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("connection closed"))
			},
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockUserServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewUserClient(mockSvc)
			err := client.UpdateRole(context.Background(), 1, "admin", "key-role")

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
