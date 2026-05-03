package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/repository/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
)

func TestClientProfileUseCase_CreateProfile(t *testing.T) {
	type mockInit func(m *mocks.MockClientProfileRepository)

	ctx := context.Background()
	accountID := int64(1)
	idemKey := "some-key"

	tests := []struct {
		name          string
		mockInit      mockInit
		expectedError error
	}{
		{
			name: "Успешное создание профиля",
			mockInit: func(m *mocks.MockClientProfileRepository) {
				m.EXPECT().Create(gomock.Any(), accountID, idemKey).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Ошибка репозитория",
			mockInit: func(m *mocks.MockClientProfileRepository) {
				m.EXPECT().Create(gomock.Any(), accountID, idemKey).Return(errors.New("db error"))
			},
			expectedError: errutil.Internal("failed to create client profile in db", errors.New("db error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockClientProfileRepository(ctrl)
			tt.mockInit(repo)

			uc := NewClientProfileUseCase(repo)
			err := uc.CreateProfile(ctx, accountID, idemKey)

			if tt.expectedError != nil {
				assert.Error(t, err)
				// Проверяем тип ошибки и содержимое
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClientProfileUseCase_GetByAccountID(t *testing.T) {
	type mockInit func(m *mocks.MockClientProfileRepository)

	ctx := context.Background()
	accountID := int64(42)
	expectedProfile := domain.ClientProfile{
		AccountID:    accountID,
		BonusBalance: 500,
	}

	tests := []struct {
		name            string
		mockInit        mockInit
		expectedProfile domain.ClientProfile
		expectedError   error
	}{
		{
			name: "Успешное получение профиля",
			mockInit: func(m *mocks.MockClientProfileRepository) {
				m.EXPECT().GetByAccountID(gomock.Any(), accountID).Return(expectedProfile, nil)
			},
			expectedProfile: expectedProfile,
			expectedError:   nil,
		},
		{
			name: "Профиль не найден",
			mockInit: func(m *mocks.MockClientProfileRepository) {
				m.EXPECT().GetByAccountID(gomock.Any(), accountID).Return(domain.ClientProfile{}, domain.ErrUserNotFound)
			},
			expectedError: errutil.New("PROFILE_NOT_FOUND", "client profile not found", codes.NotFound),
		},
		{
			name: "Внутренняя ошибка репозитория",
			mockInit: func(m *mocks.MockClientProfileRepository) {
				m.EXPECT().GetByAccountID(gomock.Any(), accountID).Return(domain.ClientProfile{}, errors.New("connection failed"))
			},
			expectedError: errutil.Internal("failed to get client profile from db", errors.New("connection failed")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockClientProfileRepository(ctrl)
			tt.mockInit(repo)

			uc := NewClientProfileUseCase(repo)
			profile, err := uc.GetByAccountID(ctx, accountID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())

				// Если это доменная ошибка с кодом, проверяем статус
				if coder, ok := err.(interface{ GRPCStatus() codes.Code }); ok {
					if expectedCoder, ok := tt.expectedError.(interface{ GRPCStatus() codes.Code }); ok {
						assert.Equal(t, expectedCoder.GRPCStatus(), coder.GRPCStatus())
					}
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedProfile, profile)
			}
		})
	}
}
