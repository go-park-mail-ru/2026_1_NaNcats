package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/usecase/mocks"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/user"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
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

			h := NewUserHandler(userUC, nil, nil)
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

			h := NewUserHandler(nil, clientUC, nil)
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

			h := NewUserHandler(userUC, nil, nil)
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

			h := NewUserHandler(userUC, nil, nil)
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

			h := NewUserHandler(userUC, nil, nil)
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

			h := NewUserHandler(userUC, nil, nil)
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

			h := NewUserHandler(userUC, nil, nil)
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

			h := NewUserHandler(userUC, nil, nil)
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

			h := NewUserHandler(userUC, clientUC, nil)
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

			h := NewUserHandler(userUC, nil, nil)
			resp, err := h.UpdateUserRole(context.Background(), tt.req)

			st, _ := status.FromError(err)
			assert.Equal(t, tt.expectedCode, st.Code())
			if tt.expectedCode == codes.OK {
				assert.Equal(t, &emptypb.Empty{}, resp)
			}
		})
	}
}

func TestUserHandler_GetUsersByIDs(t *testing.T) {
	type mockInit func(u *mocks.MockUserUseCase)
	req := &pb.GetUsersByIDsRequest{UserIds: []int64{1, 2}}
	usersMap := map[int64]domain.User{
		1: {ID: 1, Name: "User1", Email: "user1@mail.ru"},
		2: {ID: 2, Name: "User2", Email: "user2@mail.ru"},
	}

	tests := []struct {
		name         string
		req          *pb.GetUsersByIDsRequest
		mockInit     mockInit
		expectedCode codes.Code
		expectedMap  map[int64]*pb.User
	}{
		{
			name: "Успешное получение списка пользователей",
			req:  req,
			mockInit: func(u *mocks.MockUserUseCase) {
				u.EXPECT().GetUsersByIDs(gomock.Any(), req.UserIds).Return(usersMap, nil)
			},
			expectedCode: codes.OK,
			expectedMap: map[int64]*pb.User{
				1: {Id: 1, Name: "User1", Email: "user1@mail.ru"},
				2: {Id: 2, Name: "User2", Email: "user2@mail.ru"},
			},
		},
		{
			name: "Ошибка при получении пользователей",
			req:  req,
			mockInit: func(u *mocks.MockUserUseCase) {
				u.EXPECT().GetUsersByIDs(gomock.Any(), req.UserIds).Return(nil, errors.New("db error"))
			},
			expectedCode: codes.Internal,
			expectedMap:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userUC := mocks.NewMockUserUseCase(ctrl)
			tt.mockInit(userUC)

			h := NewUserHandler(userUC, nil, nil)
			resp, err := h.GetUsersByIDs(context.Background(), tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMap, resp.Users)
			} else {
				st, _ := status.FromError(err)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestUserHandler_ResolvePublicID(t *testing.T) {
	type mockInit func(u *mocks.MockUserUseCase)
	pubID := "public-uuid-123"
	req := &pb.ResolvePublicIDRequest{PublicId: pubID}

	tests := []struct {
		name         string
		req          *pb.ResolvePublicIDRequest
		mockInit     mockInit
		expectedCode codes.Code
		expectedID   int64
	}{
		{
			name: "Успешный резолв публичного ID",
			req:  req,
			mockInit: func(u *mocks.MockUserUseCase) {
				u.EXPECT().GetByPublicID(gomock.Any(), pubID).Return(domain.User{ID: 42, PublicID: pubID}, nil)
			},
			expectedCode: codes.OK,
			expectedID:   42,
		},
		{
			name: "Пользователь не найден",
			req:  req,
			mockInit: func(u *mocks.MockUserUseCase) {
				u.EXPECT().GetByPublicID(gomock.Any(), pubID).Return(domain.User{}, domain.ErrUserNotFound)
			},
			expectedCode: codes.NotFound,
			expectedID:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userUC := mocks.NewMockUserUseCase(ctrl)
			tt.mockInit(userUC)

			h := NewUserHandler(userUC, nil, nil)
			resp, err := h.ResolvePublicID(context.Background(), tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, resp.UserId)
			} else {
				st, _ := status.FromError(err)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestUserHandler_ListAchievements(t *testing.T) {
	type mockInit func(a *mocks.MockAchievementUseCase)
	req := &emptypb.Empty{}
	domainList := []domain.Achievement{
		{ID: 1, Code: "ACH1", Title: "Title1", Description: "Desc1", Icon: "icon1.png", SortOrder: 1},
	}

	tests := []struct {
		name         string
		req          *emptypb.Empty
		mockInit     mockInit
		expectedCode codes.Code
		expectedResp *pb.ListAchievementsResponse
	}{
		{
			name: "Успешное получение списка достижений",
			req:  req,
			mockInit: func(a *mocks.MockAchievementUseCase) {
				a.EXPECT().ListAll(gomock.Any()).Return(domainList, nil)
			},
			expectedCode: codes.OK,
			expectedResp: &pb.ListAchievementsResponse{
				Achievements: []*pb.Achievement{
					{Id: 1, Code: "ACH1", Title: "Title1", Description: "Desc1", Icon: "icon1.png", SortOrder: 1},
				},
			},
		},
		{
			name: "Ошибка при получении списка достижений",
			req:  req,
			mockInit: func(a *mocks.MockAchievementUseCase) {
				a.EXPECT().ListAll(gomock.Any()).Return(nil, errors.New("db error"))
			},
			expectedCode: codes.Internal,
			expectedResp: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			achieveUC := mocks.NewMockAchievementUseCase(ctrl)
			tt.mockInit(achieveUC)

			h := NewUserHandler(nil, nil, achieveUC)
			resp, err := h.ListAchievements(context.Background(), tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			} else {
				st, _ := status.FromError(err)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestUserHandler_GetUserAchievements(t *testing.T) {
	type mockInit func(a *mocks.MockAchievementUseCase)
	req := &pb.GetUserAchievementsRequest{UserId: 1}
	awardedTime := time.Now()
	domainList := []domain.UserAchievement{
		{AchievementID: 42, AwardedAt: awardedTime},
	}

	tests := []struct {
		name         string
		req          *pb.GetUserAchievementsRequest
		mockInit     mockInit
		expectedCode codes.Code
		expectedResp *pb.GetUserAchievementsResponse
	}{
		{
			name: "Успешное получение достижений пользователя",
			req:  req,
			mockInit: func(a *mocks.MockAchievementUseCase) {
				a.EXPECT().ListForUser(gomock.Any(), int64(1)).Return(domainList, nil)
			},
			expectedCode: codes.OK,
			expectedResp: &pb.GetUserAchievementsResponse{
				Achievements: []*pb.UserAchievement{
					{AchievementId: 42, AwardedAt: timestamppb.New(awardedTime)},
				},
			},
		},
		{
			name: "Ошибка при получении достижений пользователя",
			req:  req,
			mockInit: func(a *mocks.MockAchievementUseCase) {
				a.EXPECT().ListForUser(gomock.Any(), int64(1)).Return(nil, errors.New("db error"))
			},
			expectedCode: codes.Internal,
			expectedResp: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			achieveUC := mocks.NewMockAchievementUseCase(ctrl)
			tt.mockInit(achieveUC)

			h := NewUserHandler(nil, nil, achieveUC)
			resp, err := h.GetUserAchievements(context.Background(), tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp.Achievements[0].AchievementId, resp.Achievements[0].AchievementId)
				assert.Equal(t, tt.expectedResp.Achievements[0].AwardedAt.Seconds, resp.Achievements[0].AwardedAt.Seconds)
			} else {
				st, _ := status.FromError(err)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestUserHandler_ActivateStreakFreeze(t *testing.T) {
	type mockInit func(c *mocks.MockClientProfileUseCase)
	req := &pb.ActivateStreakFreezeRequest{UserId: 42}

	tests := []struct {
		name         string
		req          *pb.ActivateStreakFreezeRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешная активация заморозки",
			req:  req,
			mockInit: func(c *mocks.MockClientProfileUseCase) {
				c.EXPECT().ActivateStreakFreeze(gomock.Any(), int64(42)).Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка активации заморозки",
			req:  req,
			mockInit: func(c *mocks.MockClientProfileUseCase) {
				c.EXPECT().ActivateStreakFreeze(gomock.Any(), int64(42)).Return(errors.New("db error"))
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

			h := NewUserHandler(nil, clientUC, nil)
			resp, err := h.ActivateStreakFreeze(context.Background(), tt.req)

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

func TestUserHandler_IncrementStreak(t *testing.T) {
	type mockInit func(c *mocks.MockClientProfileUseCase)
	req := &pb.IncrementStreakRequest{UserId: 42}

	tests := []struct {
		name         string
		req          *pb.IncrementStreakRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешное увеличение стрика",
			req:  req,
			mockInit: func(c *mocks.MockClientProfileUseCase) {
				c.EXPECT().IncrementStreak(gomock.Any(), int64(42)).Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка при увеличении стрика",
			req:  req,
			mockInit: func(c *mocks.MockClientProfileUseCase) {
				c.EXPECT().IncrementStreak(gomock.Any(), int64(42)).Return(errors.New("db error"))
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

			h := NewUserHandler(nil, clientUC, nil)
			resp, err := h.IncrementStreak(context.Background(), tt.req)

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

func TestUserHandler_OnWheelSpin(t *testing.T) {
	type mockInit func(a *mocks.MockAchievementUseCase)
	wonCode := "LUCKY"
	req := &pb.OnWheelSpinRequest{UserId: 42, WonAchievementCode: &wonCode}

	tests := []struct {
		name         string
		req          *pb.OnWheelSpinRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешная обработка вращения колеса",
			req:  req,
			mockInit: func(a *mocks.MockAchievementUseCase) {
				a.EXPECT().OnWheelSpin(gomock.Any(), int64(42), "LUCKY").Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Успешная обработка вращения колеса без кода",
			req:  &pb.OnWheelSpinRequest{UserId: 42},
			mockInit: func(a *mocks.MockAchievementUseCase) {
				a.EXPECT().OnWheelSpin(gomock.Any(), int64(42), "").Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка при обработке вращения колеса",
			req:  req,
			mockInit: func(a *mocks.MockAchievementUseCase) {
				a.EXPECT().OnWheelSpin(gomock.Any(), int64(42), "LUCKY").Return(errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			achieveUC := mocks.NewMockAchievementUseCase(ctrl)
			tt.mockInit(achieveUC)

			h := NewUserHandler(nil, nil, achieveUC)
			resp, err := h.OnWheelSpin(context.Background(), tt.req)

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

func TestUserHandler_OnWordleResult(t *testing.T) {
	type mockInit func(a *mocks.MockAchievementUseCase)
	req := &pb.OnWordleResultRequest{UserId: 42, IsWin: true, TotalWins: 10, CurrentStreak: 5}

	tests := []struct {
		name         string
		req          *pb.OnWordleResultRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешная обработка результата wordle",
			req:  req,
			mockInit: func(a *mocks.MockAchievementUseCase) {
				a.EXPECT().OnWordleResult(gomock.Any(), int64(42), true, int32(10), int32(5)).Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка при обработке результата wordle",
			req:  req,
			mockInit: func(a *mocks.MockAchievementUseCase) {
				a.EXPECT().OnWordleResult(gomock.Any(), int64(42), true, int32(10), int32(5)).Return(errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			achieveUC := mocks.NewMockAchievementUseCase(ctrl)
			tt.mockInit(achieveUC)

			h := NewUserHandler(nil, nil, achieveUC)
			resp, err := h.OnWordleResult(context.Background(), tt.req)

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

func TestUserHandler_ClaimWheelSpin(t *testing.T) {
	type mockInit func(c *mocks.MockClientProfileUseCase)
	req := &pb.ClaimWheelSpinRequest{UserId: 42}

	tests := []struct {
		name         string
		req          *pb.ClaimWheelSpinRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешное применение вращения колеса",
			req:  req,
			mockInit: func(c *mocks.MockClientProfileUseCase) {
				c.EXPECT().ClaimWheelSpin(gomock.Any(), int64(42)).Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка при применении вращения",
			req:  req,
			mockInit: func(c *mocks.MockClientProfileUseCase) {
				c.EXPECT().ClaimWheelSpin(gomock.Any(), int64(42)).Return(errors.New("db error"))
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

			h := NewUserHandler(nil, clientUC, nil)
			resp, err := h.ClaimWheelSpin(context.Background(), tt.req)

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

func TestUserHandler_ResetWheelSpinCooldown(t *testing.T) {
	type mockInit func(c *mocks.MockClientProfileUseCase)
	req := &pb.ResetWheelSpinCooldownRequest{UserId: 42}

	tests := []struct {
		name         string
		req          *pb.ResetWheelSpinCooldownRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешный сброс кулдауна",
			req:  req,
			mockInit: func(c *mocks.MockClientProfileUseCase) {
				c.EXPECT().ResetWheelSpinCooldown(gomock.Any(), int64(42)).Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка при сбросе кулдауна",
			req:  req,
			mockInit: func(c *mocks.MockClientProfileUseCase) {
				c.EXPECT().ResetWheelSpinCooldown(gomock.Any(), int64(42)).Return(errors.New("db error"))
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

			h := NewUserHandler(nil, clientUC, nil)
			resp, err := h.ResetWheelSpinCooldown(context.Background(), tt.req)

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
