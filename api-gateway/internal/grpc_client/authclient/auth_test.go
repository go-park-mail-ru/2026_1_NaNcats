package authclient

import (
	"context"
	"errors"
	"testing"

	pbAuth "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/auth"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/auth/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

//go:generate mockgen -destination=../../../../shared/proto/auth/mocks/auth_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/auth AuthServiceClient
func TestAuthClient_IssueSession(t *testing.T) {
	type mockBehavior func(m *mocks.MockAuthServiceClient)

	expectedSession := &pbAuth.Session{
		Id:        "session-123",
		UserId:    1,
		Role:      "user",
		UserAgent: "Mozilla",
	}

	tests := []struct {
		name         string
		userID       int64
		role         string
		userAgent    string
		mockBehavior mockBehavior
		expectedRes  *pbAuth.Session
		expectedErr  error
	}{
		{
			name:      "Успешная выдача сессии",
			userID:    1,
			role:      "user",
			userAgent: "Mozilla",
			mockBehavior: func(m *mocks.MockAuthServiceClient) {
				m.EXPECT().IssueSession(gomock.Any(), &pbAuth.IssueSessionRequest{
					UserId:    1,
					Role:      "user",
					UserAgent: "Mozilla",
				}).Return(&pbAuth.AuthResponse{Session: expectedSession}, nil)
			},
			expectedRes: expectedSession,
			expectedErr: nil,
		},
		{
			name:      "Ошибка при выдаче (Internal)",
			userID:    1,
			role:      "user",
			userAgent: "Mozilla",
			mockBehavior: func(m *mocks.MockAuthServiceClient) {
				m.EXPECT().IssueSession(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.Internal, "db error"))
			},
			expectedRes: nil,
			expectedErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGRPCClient := mocks.NewMockAuthServiceClient(ctrl)
			tt.mockBehavior(mockGRPCClient)

			client := NewAuthClient(mockGRPCClient)
			res, err := client.IssueSession(context.Background(), tt.userID, tt.role, tt.userAgent)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes.Id, res.Id)
			}
		})
	}
}

func TestAuthClient_Login(t *testing.T) {
	type mockBehavior func(m *mocks.MockAuthServiceClient)

	expectedSession := &pbAuth.Session{
		Id:        "session-123",
		UserId:    1,
		Role:      "user",
		UserAgent: "Mozilla",
	}

	tests := []struct {
		name         string
		email        string
		password     string
		userAgent    string
		mockBehavior mockBehavior
		expectedRes  *pbAuth.Session
		expectedErr  error
	}{
		{
			name:      "Успешный логин",
			email:     "test@mail.ru",
			password:  "pass123",
			userAgent: "Mozilla",
			mockBehavior: func(m *mocks.MockAuthServiceClient) {
				m.EXPECT().Login(gomock.Any(), &pbAuth.LoginRequest{
					Email:     "test@mail.ru",
					Password:  "pass123",
					UserAgent: "Mozilla",
				}).Return(&pbAuth.AuthResponse{Session: expectedSession}, nil)
			},
			expectedRes: expectedSession,
			expectedErr: nil,
		},
		{
			name:      "Ошибка: неверные учетные данные (Unauthenticated)",
			email:     "test@mail.ru",
			password:  "wrong",
			userAgent: "Mozilla",
			mockBehavior: func(m *mocks.MockAuthServiceClient) {
				m.EXPECT().Login(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.Unauthenticated, "invalid pass"))
			},
			expectedRes: nil,
			expectedErr: ErrInvalidCredentials,
		},
		{
			name:      "Внутренняя ошибка сервиса",
			email:     "test@mail.ru",
			password:  "pass123",
			userAgent: "Mozilla",
			mockBehavior: func(m *mocks.MockAuthServiceClient) {
				m.EXPECT().Login(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.Internal, "db down"))
			},
			expectedRes: nil,
			expectedErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGRPCClient := mocks.NewMockAuthServiceClient(ctrl)
			tt.mockBehavior(mockGRPCClient)

			client := NewAuthClient(mockGRPCClient)
			res, err := client.Login(context.Background(), tt.email, tt.password, tt.userAgent)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes.Id, res.Id)
			}
		})
	}
}

func TestAuthClient_Logout(t *testing.T) {
	type mockBehavior func(m *mocks.MockAuthServiceClient)

	tests := []struct {
		name         string
		sessionID    string
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name:      "Успешный логаут",
			sessionID: "sess-123",
			mockBehavior: func(m *mocks.MockAuthServiceClient) {
				m.EXPECT().Logout(gomock.Any(), &pbAuth.LogoutRequest{
					SessionId: "sess-123",
				}).Return(&emptypb.Empty{}, nil)
			},
			expectedErr: nil,
		},
		{
			name:      "Ошибка логаута",
			sessionID: "sess-err",
			mockBehavior: func(m *mocks.MockAuthServiceClient) {
				m.EXPECT().Logout(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("some error"))
			},
			expectedErr: errors.New("some error"), // Отдаем как есть
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGRPCClient := mocks.NewMockAuthServiceClient(ctrl)
			tt.mockBehavior(mockGRPCClient)

			client := NewAuthClient(mockGRPCClient)
			err := client.Logout(context.Background(), tt.sessionID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthClient_CheckSession(t *testing.T) {
	type mockBehavior func(m *mocks.MockAuthServiceClient)

	tests := []struct {
		name         string
		sessionID    string
		mockBehavior mockBehavior
		expectedID   int64
		expectedRole string
		expectedErr  error
	}{
		{
			name:      "Успешная проверка сессии",
			sessionID: "sess-123",
			mockBehavior: func(m *mocks.MockAuthServiceClient) {
				m.EXPECT().CheckSession(gomock.Any(), &pbAuth.CheckSessionRequest{
					SessionId: "sess-123",
				}).Return(&pbAuth.CheckSessionResponse{
					UserId: 42,
					Role:   "admin",
				}, nil)
			},
			expectedID:   42,
			expectedRole: "admin",
			expectedErr:  nil,
		},
		{
			name:      "Сессия не найдена (NotFound)",
			sessionID: "sess-404",
			mockBehavior: func(m *mocks.MockAuthServiceClient) {
				m.EXPECT().CheckSession(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectedID:   0,
			expectedRole: "",
			expectedErr:  ErrSessionNotFound,
		},
		{
			name:      "Сессия истекла (Unauthenticated)",
			sessionID: "sess-exp",
			mockBehavior: func(m *mocks.MockAuthServiceClient) {
				m.EXPECT().CheckSession(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.Unauthenticated, "expired"))
			},
			expectedID:   0,
			expectedRole: "",
			expectedErr:  ErrSessionNotFound,
		},
		{
			name:      "Внутренняя ошибка",
			sessionID: "sess-err",
			mockBehavior: func(m *mocks.MockAuthServiceClient) {
				m.EXPECT().CheckSession(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.Internal, "db error"))
			},
			expectedID:   0,
			expectedRole: "",
			expectedErr:  ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGRPCClient := mocks.NewMockAuthServiceClient(ctrl)
			tt.mockBehavior(mockGRPCClient)

			client := NewAuthClient(mockGRPCClient)
			id, role, err := client.CheckSession(context.Background(), tt.sessionID)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
				assert.Equal(t, tt.expectedRole, role)
			}
		})
	}
}

func TestAuthClient_SetCSRF(t *testing.T) {
	type mockBehavior func(m *mocks.MockAuthServiceClient)

	tests := []struct {
		name          string
		sessionID     string
		mockBehavior  mockBehavior
		expectedToken string
		expectedErr   error
	}{
		{
			name:      "Успешная генерация токена",
			sessionID: "sess-123",
			mockBehavior: func(m *mocks.MockAuthServiceClient) {
				m.EXPECT().SetCSRF(gomock.Any(), &pbAuth.CSRFRequest{
					SessionId: "sess-123",
				}).Return(&pbAuth.CSRFResponse{Token: "token-123"}, nil)
			},
			expectedToken: "token-123",
			expectedErr:   nil,
		},
		{
			name:      "Внутренняя ошибка при генерации",
			sessionID: "sess-err",
			mockBehavior: func(m *mocks.MockAuthServiceClient) {
				m.EXPECT().SetCSRF(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.Internal, "redis error"))
			},
			expectedToken: "",
			expectedErr:   ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGRPCClient := mocks.NewMockAuthServiceClient(ctrl)
			tt.mockBehavior(mockGRPCClient)

			client := NewAuthClient(mockGRPCClient)
			token, err := client.SetCSRF(context.Background(), tt.sessionID)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken, token)
			}
		})
	}
}

func TestAuthClient_GetCSRF(t *testing.T) {
	type mockBehavior func(m *mocks.MockAuthServiceClient)

	tests := []struct {
		name          string
		sessionID     string
		mockBehavior  mockBehavior
		expectedToken string
		expectedErr   error
	}{
		{
			name:      "Успешное получение токена",
			sessionID: "sess-123",
			mockBehavior: func(m *mocks.MockAuthServiceClient) {
				m.EXPECT().GetCSRF(gomock.Any(), &pbAuth.CSRFRequest{
					SessionId: "sess-123",
				}).Return(&pbAuth.CSRFResponse{Token: "token-123"}, nil)
			},
			expectedToken: "token-123",
			expectedErr:   nil,
		},
		{
			name:      "Токен не найден (NotFound)",
			sessionID: "sess-404",
			mockBehavior: func(m *mocks.MockAuthServiceClient) {
				m.EXPECT().GetCSRF(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectedToken: "",
			expectedErr:   ErrSessionNotFound,
		},
		{
			name:      "Внутренняя ошибка",
			sessionID: "sess-err",
			mockBehavior: func(m *mocks.MockAuthServiceClient) {
				m.EXPECT().GetCSRF(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.Internal, "redis down"))
			},
			expectedToken: "",
			expectedErr:   ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGRPCClient := mocks.NewMockAuthServiceClient(ctrl)
			tt.mockBehavior(mockGRPCClient)

			client := NewAuthClient(mockGRPCClient)
			token, err := client.GetCSRF(context.Background(), tt.sessionID)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken, token)
			}
		})
	}
}
