package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/auth/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/auth/internal/usecase/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/grpcutil"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAuthHandler_IssueSession(t *testing.T) {
	type mockInit func(m *mocks.MockAuthUseCase)

	req := &pb.IssueSessionRequest{
		UserId:    1,
		Role:      "user",
		UserAgent: "Mozilla",
	}

	sessionID := uuid.New()
	expectedSession := domain.Session{
		ID:        sessionID,
		UserID:    1,
		Role:      "user",
		UserAgent: "Mozilla",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	tests := []struct {
		name         string
		req          *pb.IssueSessionRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешная выдача сессии",
			req:  req,
			mockInit: func(m *mocks.MockAuthUseCase) {
				m.EXPECT().IssueSession(gomock.Any(), req.UserId, req.Role, req.UserAgent).
					Return(expectedSession, nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка UseCase при выдаче сессии",
			req:  req,
			mockInit: func(m *mocks.MockAuthUseCase) {
				m.EXPECT().IssueSession(gomock.Any(), req.UserId, req.Role, req.UserAgent).
					Return(domain.Session{}, errors.New("internal issue error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockAuthUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewAuthHandler(mockUC)
			resp, err := handler.IssueSession(context.Background(), tt.req)

			grpcErr := grpcutil.ToGRPCError(err)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, grpcErr)
				require.NotNil(t, resp)
				assert.Equal(t, sessionID.String(), resp.Session.Id)
				assert.Equal(t, int64(1), resp.Session.UserId)
			} else {
				st, ok := status.FromError(grpcErr)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	type mockInit func(m *mocks.MockAuthUseCase)

	req := &pb.LoginRequest{
		Email:     "test@mail.ru",
		Password:  "password123",
		UserAgent: "Chrome",
	}

	sessionID := uuid.New()
	expectedSession := domain.Session{
		ID:        sessionID,
		UserID:    2,
		Role:      "admin",
		UserAgent: "Chrome",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	tests := []struct {
		name         string
		req          *pb.LoginRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешный логин",
			req:  req,
			mockInit: func(m *mocks.MockAuthUseCase) {
				m.EXPECT().Login(gomock.Any(), req.Email, req.Password, req.UserAgent).
					Return(expectedSession, nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка авторизации (неверный пароль)",
			req:  req,
			mockInit: func(m *mocks.MockAuthUseCase) {
				m.EXPECT().Login(gomock.Any(), req.Email, req.Password, req.UserAgent).
					Return(domain.Session{}, domain.ErrSessionExpired)
			},
			expectedCode: codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockAuthUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewAuthHandler(mockUC)
			resp, err := handler.Login(context.Background(), tt.req)

			grpcErr := grpcutil.ToGRPCError(err)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, grpcErr)
				require.NotNil(t, resp)
				assert.Equal(t, sessionID.String(), resp.Session.Id)
			} else {
				st, _ := status.FromError(grpcErr)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	type mockInit func(m *mocks.MockAuthUseCase)

	validUUID := uuid.New()

	tests := []struct {
		name         string
		req          *pb.LogoutRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешный логаут",
			req:  &pb.LogoutRequest{SessionId: validUUID.String()},
			mockInit: func(m *mocks.MockAuthUseCase) {
				m.EXPECT().Logout(gomock.Any(), validUUID).Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name:         "Ошибка: невалидный UUID",
			req:          &pb.LogoutRequest{SessionId: "invalid-uuid-format"},
			mockInit:     func(m *mocks.MockAuthUseCase) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Ошибка в UseCase",
			req:  &pb.LogoutRequest{SessionId: validUUID.String()},
			mockInit: func(m *mocks.MockAuthUseCase) {
				m.EXPECT().Logout(gomock.Any(), validUUID).Return(errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockAuthUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewAuthHandler(mockUC)
			_, err := handler.Logout(context.Background(), tt.req)

			grpcErr := grpcutil.ToGRPCError(err)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, grpcErr)
			} else {
				st, _ := status.FromError(grpcErr)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestAuthHandler_CheckSession(t *testing.T) {
	type mockInit func(m *mocks.MockAuthUseCase)

	validUUID := uuid.New()

	tests := []struct {
		name         string
		req          *pb.CheckSessionRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешная проверка сессии",
			req:  &pb.CheckSessionRequest{SessionId: validUUID.String()},
			mockInit: func(m *mocks.MockAuthUseCase) {
				m.EXPECT().CheckUserSession(gomock.Any(), validUUID).
					Return(int64(42), "user", nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка: невалидный UUID",
			req:  &pb.CheckSessionRequest{SessionId: "bad-uuid"},
			mockInit: func(m *mocks.MockAuthUseCase) {
			},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Сессия не найдена",
			req:  &pb.CheckSessionRequest{SessionId: validUUID.String()},
			mockInit: func(m *mocks.MockAuthUseCase) {
				m.EXPECT().CheckUserSession(gomock.Any(), validUUID).
					Return(int64(0), "", domain.ErrSessionNotFound)
			},
			expectedCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockAuthUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewAuthHandler(mockUC)
			resp, err := handler.CheckSession(context.Background(), tt.req)

			grpcErr := grpcutil.ToGRPCError(err)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, grpcErr)
				require.NotNil(t, resp)
				assert.Equal(t, int64(42), resp.UserId)
				assert.Equal(t, "user", resp.Role)
			} else {
				st, _ := status.FromError(grpcErr)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestAuthHandler_GetCSRF(t *testing.T) {
	type mockInit func(m *mocks.MockAuthUseCase)

	validUUID := uuid.New()

	tests := []struct {
		name         string
		req          *pb.CSRFRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешное получение CSRF",
			req:  &pb.CSRFRequest{SessionId: validUUID.String()},
			mockInit: func(m *mocks.MockAuthUseCase) {
				m.EXPECT().GetCSRFBySessionID(gomock.Any(), validUUID).
					Return("token-123", nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка: невалидный UUID",
			req:  &pb.CSRFRequest{SessionId: "123"},
			mockInit: func(m *mocks.MockAuthUseCase) {
			},
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockAuthUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewAuthHandler(mockUC)
			resp, err := handler.GetCSRF(context.Background(), tt.req)

			grpcErr := grpcutil.ToGRPCError(err)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, grpcErr)
				require.NotNil(t, resp)
				assert.Equal(t, "token-123", resp.Token)
			} else {
				st, _ := status.FromError(grpcErr)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestAuthHandler_SetCSRF(t *testing.T) {
	type mockInit func(m *mocks.MockAuthUseCase)

	validUUID := uuid.New()

	tests := []struct {
		name         string
		req          *pb.CSRFRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешная установка CSRF",
			req:  &pb.CSRFRequest{SessionId: validUUID.String()},
			mockInit: func(m *mocks.MockAuthUseCase) {
				m.EXPECT().SetCSRFForUser(gomock.Any(), validUUID).
					Return("new-token-123", nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка: невалидный UUID",
			req:  &pb.CSRFRequest{SessionId: "bad"},
			mockInit: func(m *mocks.MockAuthUseCase) {
			},
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockAuthUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewAuthHandler(mockUC)
			resp, err := handler.SetCSRF(context.Background(), tt.req)

			grpcErr := grpcutil.ToGRPCError(err)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, grpcErr)
				require.NotNil(t, resp)
				assert.Equal(t, "new-token-123", resp.Token)
			} else {
				st, _ := status.FromError(grpcErr)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}
