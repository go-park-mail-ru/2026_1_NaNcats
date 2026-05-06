package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/usecase/mocks"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/user"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestUserHandler_CreateUser(t *testing.T) {
	type mockInit func(u *mocks.MockUserUseCase)
	tests := []struct {
		name         string
		req          *pb.CreateUserRequest
		mockInit     mockInit
		expectedResp *pb.CreateUserResponse
		expectedCode codes.Code
	}{
		{
			name: "Успешное создание пользователя",
			req: &pb.CreateUserRequest{
				Name:           "Ivan",
				Email:          "ivan@example.com",
				Password:       "password123",
				IdempotencyKey: "idem-1",
			},
			mockInit: func(u *mocks.MockUserUseCase) {
				u.EXPECT().
					Create(gomock.Any(), domain.User{Name: "Ivan", Email: "ivan@example.com"}, "password123", "idem-1").
					Return(int64(42), nil)
			},
			expectedResp: &pb.CreateUserResponse{UserId: 42},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка: пользователь уже существует",
			req: &pb.CreateUserRequest{
				Email: "exists@example.com",
			},
			mockInit: func(u *mocks.MockUserUseCase) {
				u.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(int64(0), domain.ErrEmailAlreadyExists)
			},
			expectedCode: codes.AlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userUC := mocks.NewMockUserUseCase(ctrl)
			tt.mockInit(userUC)

			h := NewUserHandler(userUC, nil)
			resp, err := h.CreateUser(context.Background(), tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			} else {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestUserHandler_CreateClientProfile(t *testing.T) {
	type mockInit func(c *mocks.MockClientProfileUseCase)
	tests := []struct {
		name         string
		req          *pb.CreateClientProfileRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешное создание профиля",
			req: &pb.CreateClientProfileRequest{
				UserId:         42,
				IdempotencyKey: "idem-2",
			},
			mockInit: func(c *mocks.MockClientProfileUseCase) {
				c.EXPECT().
					CreateProfile(gomock.Any(), int64(42), "idem-2").
					Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка UseCase при создании профиля",
			req: &pb.CreateClientProfileRequest{
				UserId: 42,
			},
			mockInit: func(c *mocks.MockClientProfileUseCase) {
				c.EXPECT().
					CreateProfile(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(status.Error(codes.Internal, "db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			clientUC := mocks.NewMockClientProfileUseCase(ctrl)
			tt.mockInit(clientUC)

			h := NewUserHandler(nil, clientUC)
			resp, err := h.CreateClientProfile(context.Background(), tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.Equal(t, &emptypb.Empty{}, resp)
			} else {
				st, _ := status.FromError(err)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestUserHandler_UpdateProfile(t *testing.T) {
	type mockInit func(u *mocks.MockUserUseCase)
	name := "NewName"
	email := "new@example.com"

	tests := []struct {
		name         string
		req          *pb.UpdateProfileRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешное обновление профиля",
			req: &pb.UpdateProfileRequest{
				UserId:         1,
				Name:           &name,
				Email:          &email,
				IdempotencyKey: "idem-3",
			},
			mockInit: func(u *mocks.MockUserUseCase) {
				u.EXPECT().
					UpdateProfile(gomock.Any(), int64(1), &name, &email, "idem-3").
					Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка: данные для обновления не предоставлены",
			req: &pb.UpdateProfileRequest{
				UserId: 1,
			},
			mockInit: func(u *mocks.MockUserUseCase) {
				u.EXPECT().
					UpdateProfile(gomock.Any(), int64(1), nil, nil, gomock.Any()).
					Return(domain.ErrNoChangesProvided)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Ошибка: пользователь не найден",
			req: &pb.UpdateProfileRequest{
				UserId: 999,
				Name:   &name,
			},
			mockInit: func(u *mocks.MockUserUseCase) {
				u.EXPECT().
					UpdateProfile(gomock.Any(), int64(999), &name, nil, gomock.Any()).
					Return(domain.ErrUserNotFound)
			},
			expectedCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userUC := mocks.NewMockUserUseCase(ctrl)
			tt.mockInit(userUC)

			h := NewUserHandler(userUC, nil)
			resp, err := h.UpdateProfile(context.Background(), tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.Equal(t, &emptypb.Empty{}, resp)
			} else {
				st, _ := status.FromError(err)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestUserHandler_UpdateAvatar(t *testing.T) {
	type mockInit func(u *mocks.MockUserUseCase)
	imageData := []byte("fake-image-bytes")

	tests := []struct {
		name         string
		req          *pb.UpdateAvatarRequest
		mockInit     mockInit
		expectedURL  string
		expectedCode codes.Code
	}{
		{
			name: "Успешное обновление аватара",
			req: &pb.UpdateAvatarRequest{
				UserId:         1,
				ImageData:      imageData,
				IdempotencyKey: "idem-avatar-1",
			},
			mockInit: func(u *mocks.MockUserUseCase) {
				u.EXPECT().
					UpdateAvatar(gomock.Any(), int64(1), imageData, "idem-avatar-1").
					Return("http://s3.ru/new.webp", nil)
			},
			expectedURL:  "http://s3.ru/new.webp",
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка: неверный формат изображения",
			req: &pb.UpdateAvatarRequest{
				UserId:    1,
				ImageData: imageData,
			},
			mockInit: func(u *mocks.MockUserUseCase) {
				u.EXPECT().
					UpdateAvatar(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", domain.ErrInvalidImageExt)
			},
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userUC := mocks.NewMockUserUseCase(ctrl)
			tt.mockInit(userUC)

			h := NewUserHandler(userUC, nil)
			resp, err := h.UpdateAvatar(context.Background(), tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedURL, resp.AvatarUrl)
			} else {
				st, _ := status.FromError(err)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestUserHandler_DeleteAvatar(t *testing.T) {
	type mockInit func(u *mocks.MockUserUseCase)

	tests := []struct {
		name         string
		req          *pb.DeleteAvatarRequest
		mockInit     mockInit
		expectedURL  string
		expectedCode codes.Code
	}{
		{
			name: "Успешный сброс аватара",
			req: &pb.DeleteAvatarRequest{
				UserId:         1,
				IdempotencyKey: "idem-del-1",
			},
			mockInit: func(u *mocks.MockUserUseCase) {
				u.EXPECT().
					DeleteAvatar(gomock.Any(), int64(1), "idem-del-1").
					Return("http://s3.ru/default.webp", nil)
			},
			expectedURL:  "http://s3.ru/default.webp",
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка: пользователь не найден",
			req:  &pb.DeleteAvatarRequest{UserId: 404},
			mockInit: func(u *mocks.MockUserUseCase) {
				u.EXPECT().
					DeleteAvatar(gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", domain.ErrUserNotFound)
			},
			expectedCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userUC := mocks.NewMockUserUseCase(ctrl)
			tt.mockInit(userUC)

			h := NewUserHandler(userUC, nil)
			resp, err := h.DeleteAvatar(context.Background(), tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedURL, resp.DefaultAvatarUrl)
			} else {
				st, _ := status.FromError(err)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestUserHandler_GetByID(t *testing.T) {
	type mockInit func(u *mocks.MockUserUseCase)
	userData := domain.User{
		ID:        1,
		Name:      "Ivan",
		Email:     "ivan@mail.ru",
		Role:      "client",
		AvatarURL: "ava.png",
	}

	tests := []struct {
		name         string
		req          *pb.GetUserByIDRequest
		mockInit     mockInit
		expectedName string
		expectedCode codes.Code
	}{
		{
			name: "Пользователь найден",
			req:  &pb.GetUserByIDRequest{UserId: 1},
			mockInit: func(u *mocks.MockUserUseCase) {
				u.EXPECT().
					GetByID(gomock.Any(), int64(1)).
					Return(userData, nil)
			},
			expectedName: "Ivan",
			expectedCode: codes.OK,
		},
		{
			name: "Пользователь не найден",
			req:  &pb.GetUserByIDRequest{UserId: 99},
			mockInit: func(u *mocks.MockUserUseCase) {
				u.EXPECT().
					GetByID(gomock.Any(), int64(99)).
					Return(domain.User{}, domain.ErrUserNotFound)
			},
			expectedCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userUC := mocks.NewMockUserUseCase(ctrl)
			tt.mockInit(userUC)

			h := NewUserHandler(userUC, nil)
			resp, err := h.GetByID(context.Background(), tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedName, resp.User.Name)
				assert.Equal(t, int64(1), resp.User.Id)
			} else {
				st, _ := status.FromError(err)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestUserHandler_GetByEmail(t *testing.T) {
	type mockInit func(u *mocks.MockUserUseCase)
	email := "test@mail.ru"
	userData := domain.User{ID: 1, Email: email, Name: "Tester"}

	tests := []struct {
		name         string
		req          *pb.GetUserByEmailRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешное получение по email",
			req:  &pb.GetUserByEmailRequest{Email: email},
			mockInit: func(u *mocks.MockUserUseCase) {
				u.EXPECT().GetByEmail(gomock.Any(), email).Return(userData, nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Пользователь не найден",
			req:  &pb.GetUserByEmailRequest{Email: "none@mail.ru"},
			mockInit: func(u *mocks.MockUserUseCase) {
				u.EXPECT().GetByEmail(gomock.Any(), "none@mail.ru").Return(domain.User{}, domain.ErrUserNotFound)
			},
			expectedCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userUC := mocks.NewMockUserUseCase(ctrl)
			tt.mockInit(userUC)

			h := NewUserHandler(userUC, nil)
			resp, err := h.GetByEmail(context.Background(), tt.req)

			st, _ := status.FromError(err)
			assert.Equal(t, tt.expectedCode, st.Code())
			if tt.expectedCode == codes.OK {
				assert.Equal(t, email, resp.User.Email)
			}
		})
	}
}

func TestUserHandler_CheckUserExists(t *testing.T) {
	type mockInit func(u *mocks.MockUserUseCase)

	tests := []struct {
		name         string
		req          *pb.CheckUserExistsRequest
		mockInit     mockInit
		expectedEx   bool
		expectedCode codes.Code
	}{
		{
			name: "Пользователь существует",
			req:  &pb.CheckUserExistsRequest{UserId: 1},
			mockInit: func(u *mocks.MockUserUseCase) {
				u.EXPECT().Check(gomock.Any(), int64(1)).Return(true, nil)
			},
			expectedEx:   true,
			expectedCode: codes.OK,
		},
		{
			name: "Пользователь не существует",
			req:  &pb.CheckUserExistsRequest{UserId: 404},
			mockInit: func(u *mocks.MockUserUseCase) {
				u.EXPECT().Check(gomock.Any(), int64(404)).Return(false, nil)
			},
			expectedEx:   false,
			expectedCode: codes.OK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userUC := mocks.NewMockUserUseCase(ctrl)
			tt.mockInit(userUC)

			h := NewUserHandler(userUC, nil)
			resp, err := h.CheckUserExists(context.Background(), tt.req)

			st, _ := status.FromError(err)
			assert.Equal(t, tt.expectedCode, st.Code())
			if err == nil {
				assert.Equal(t, tt.expectedEx, resp.Exists)
			}
		})
	}
}

func TestUserHandler_GetUserProfile(t *testing.T) {
	// Для этого метода нужно два мока
	type mockInit func(u *mocks.MockUserUseCase, c *mocks.MockClientProfileUseCase)
	userID := int64(1)

	tests := []struct {
		name         string
		req          *pb.GetUserProfileRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешное получение полного профиля",
			req:  &pb.GetUserProfileRequest{UserId: userID},
			mockInit: func(u *mocks.MockUserUseCase, c *mocks.MockClientProfileUseCase) {
				u.EXPECT().GetByID(gomock.Any(), userID).Return(domain.User{ID: userID, Name: "Ivan"}, nil)
				c.EXPECT().GetByAccountID(gomock.Any(), userID).Return(domain.ClientProfile{AccountID: userID, BonusBalance: 100}, nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка: юзер не найден",
			req:  &pb.GetUserProfileRequest{UserId: userID},
			mockInit: func(u *mocks.MockUserUseCase, c *mocks.MockClientProfileUseCase) {
				u.EXPECT().GetByID(gomock.Any(), userID).Return(domain.User{}, domain.ErrUserNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Ошибка: профиль клиента не найден",
			req:  &pb.GetUserProfileRequest{UserId: userID},
			mockInit: func(u *mocks.MockUserUseCase, c *mocks.MockClientProfileUseCase) {
				u.EXPECT().GetByID(gomock.Any(), userID).Return(domain.User{ID: userID}, nil)
				c.EXPECT().GetByAccountID(gomock.Any(), userID).Return(domain.ClientProfile{}, domain.ErrUserNotFound)
			},
			expectedCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userUC := mocks.NewMockUserUseCase(ctrl)
			clientUC := mocks.NewMockClientProfileUseCase(ctrl)
			tt.mockInit(userUC, clientUC)

			h := NewUserHandler(userUC, clientUC)
			resp, err := h.GetUserProfile(context.Background(), tt.req)

			st, _ := status.FromError(err)
			assert.Equal(t, tt.expectedCode, st.Code())
			if tt.expectedCode == codes.OK {
				assert.NotNil(t, resp.User)
				assert.NotNil(t, resp.Profile)
			}
		})
	}
}

func TestUserHandler_UpdateUserRole(t *testing.T) {
	type mockInit func(u *mocks.MockUserUseCase)
	userID := int64(42)
	newRole := "admin"
	idemKey := "role-idem"

	tests := []struct {
		name         string
		req          *pb.UpdateUserRoleRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешное обновление роли",
			req:  &pb.UpdateUserRoleRequest{UserId: userID, NewRole: newRole, IdempotencyKey: idemKey},
			mockInit: func(u *mocks.MockUserUseCase) {
				u.EXPECT().UpdateRole(gomock.Any(), userID, newRole, idemKey).Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка: неверная роль (UseCase)",
			req:  &pb.UpdateUserRoleRequest{UserId: userID, NewRole: "god"},
			mockInit: func(u *mocks.MockUserUseCase) {
				u.EXPECT().UpdateRole(gomock.Any(), userID, "god", gomock.Any()).Return(domain.ErrInvalidInput)
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Внутренняя ошибка сервера",
			req:  &pb.UpdateUserRoleRequest{UserId: userID, NewRole: newRole},
			mockInit: func(u *mocks.MockUserUseCase) {
				u.EXPECT().UpdateRole(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("db crash"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userUC := mocks.NewMockUserUseCase(ctrl)
			tt.mockInit(userUC)

			h := NewUserHandler(userUC, nil)
			resp, err := h.UpdateUserRole(context.Background(), tt.req)

			st, _ := status.FromError(err)
			assert.Equal(t, tt.expectedCode, st.Code())
			if tt.expectedCode == codes.OK {
				assert.Equal(t, &emptypb.Empty{}, resp)
			}
		})
	}
}
