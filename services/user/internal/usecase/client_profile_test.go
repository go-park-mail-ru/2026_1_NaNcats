package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

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

			uc := NewClientProfileUseCase(repo, nil, nil)
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
	ctx := context.Background()
	accountID := int64(42)
	now := time.Now()
	twoWeeksAgo := now.AddDate(0, 0, -14)

	expectedProfileSuccess := domain.ClientProfile{
		AccountID:    accountID,
		BonusBalance: 500,
	}

	profileWithBrokenStreakNoFreeze := domain.ClientProfile{
		AccountID:          accountID,
		BonusBalance:       500,
		StreakCount:        5,
		StreakFreezeActive: false,
		LastOrderDate:      &twoWeeksAgo,
	}

	profileWithBrokenStreakWithFreeze := domain.ClientProfile{
		AccountID:          accountID,
		BonusBalance:       500,
		StreakCount:        5,
		StreakFreezeActive: true,
		LastOrderDate:      &twoWeeksAgo,
	}

	tests := []struct {
		name         string
		mockBehavior func(m *mocks.MockClientProfileRepository)
		want         domain.ClientProfile
		wantErr      error
	}{
		{
			name: "Успешное получение профиля без синхронизации стрика",
			mockBehavior: func(m *mocks.MockClientProfileRepository) {
				m.EXPECT().GetByAccountID(gomock.Any(), accountID).Return(expectedProfileSuccess, nil)
			},
			want:    expectedProfileSuccess,
			wantErr: nil,
		},
		{
			name: "Профиль не найден",
			mockBehavior: func(m *mocks.MockClientProfileRepository) {
				m.EXPECT().GetByAccountID(gomock.Any(), accountID).Return(domain.ClientProfile{}, domain.ErrUserNotFound)
			},
			want:    domain.ClientProfile{},
			wantErr: errutil.New("PROFILE_NOT_FOUND", "client profile not found", codes.NotFound),
		},
		{
			name: "Внутренняя ошибка базы данных при получении",
			mockBehavior: func(m *mocks.MockClientProfileRepository) {
				m.EXPECT().GetByAccountID(gomock.Any(), accountID).Return(domain.ClientProfile{}, errors.New("db error"))
			},
			want:    domain.ClientProfile{},
			wantErr: errutil.Internal("failed to get client profile from db", errors.New("db error")),
		},
		{
			name: "Успешный сброс стрика без заморозки",
			mockBehavior: func(m *mocks.MockClientProfileRepository) {
				m.EXPECT().GetByAccountID(gomock.Any(), accountID).Return(profileWithBrokenStreakNoFreeze, nil)
				m.EXPECT().ResetStreak(gomock.Any(), accountID).Return(nil)
			},
			want: domain.ClientProfile{
				AccountID:          accountID,
				BonusBalance:       500,
				StreakCount:        0,
				StreakFreezeActive: false,
				LastOrderDate:      nil,
			},
			wantErr: nil,
		},
		{
			name: "Ошибка репозитория при сбросе стрика",
			mockBehavior: func(m *mocks.MockClientProfileRepository) {
				m.EXPECT().GetByAccountID(gomock.Any(), accountID).Return(profileWithBrokenStreakNoFreeze, nil)
				m.EXPECT().ResetStreak(gomock.Any(), accountID).Return(errors.New("reset error"))
			},
			want:    domain.ClientProfile{},
			wantErr: errutil.Internal("failed to sync user streak", errors.New("reset error")),
		},
		{
			name: "Ошибка репозитория при обновлении заморозки",
			mockBehavior: func(m *mocks.MockClientProfileRepository) {
				m.EXPECT().GetByAccountID(gomock.Any(), accountID).Return(profileWithBrokenStreakWithFreeze, nil)
				m.EXPECT().UpdateStreakFreeze(gomock.Any(), accountID, false).Return(errors.New("freeze error"))
			},
			want:    domain.ClientProfile{},
			wantErr: errutil.Internal("failed to sync user streak", errors.New("freeze error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockClientProfileRepository(ctrl)
			tt.mockBehavior(repo)

			uc := NewClientProfileUseCase(repo, nil, nil)
			profile, err := uc.GetByAccountID(ctx, accountID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
				if coder, ok := err.(interface{ GRPCStatus() codes.Code }); ok {
					if expectedCoder, ok := tt.wantErr.(interface{ GRPCStatus() codes.Code }); ok {
						assert.Equal(t, expectedCoder.GRPCStatus(), coder.GRPCStatus())
					}
				}
			} else {
				assert.NoError(t, err)
				if tt.want.LastOrderDate == nil {
					assert.Nil(t, profile.LastOrderDate)
				} else {
					assert.NotNil(t, profile.LastOrderDate)
				}
				profile.LastOrderDate = tt.want.LastOrderDate
				assert.Equal(t, tt.want, profile)
			}
		})
	}
}

func TestClientProfileUseCase_ActivateStreakFreeze(t *testing.T) {
	ctx := context.Background()
	accountID := int64(10)
	now := time.Now()
	twoWeeksAgo := now.AddDate(0, 0, -14)

	tests := []struct {
		name         string
		mockBehavior func(m *mocks.MockClientProfileRepository)
		wantErr      error
	}{
		{
			name: "Успешная активация заморозки без нужды в синхронизации",
			mockBehavior: func(m *mocks.MockClientProfileRepository) {
				m.EXPECT().GetByAccountID(gomock.Any(), accountID).Return(domain.ClientProfile{AccountID: accountID}, nil)
				m.EXPECT().UpdateStreakFreeze(gomock.Any(), accountID, true).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "Ошибка при получении профиля",
			mockBehavior: func(m *mocks.MockClientProfileRepository) {
				m.EXPECT().GetByAccountID(gomock.Any(), accountID).Return(domain.ClientProfile{}, errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
		{
			name: "Ошибка при синхронизации стрика перед активацией",
			mockBehavior: func(m *mocks.MockClientProfileRepository) {
				m.EXPECT().GetByAccountID(gomock.Any(), accountID).Return(domain.ClientProfile{
					AccountID:     accountID,
					LastOrderDate: &twoWeeksAgo,
				}, nil)
				m.EXPECT().ResetStreak(gomock.Any(), accountID).Return(errors.New("reset error"))
			},
			wantErr: errors.New("reset error"),
		},
		{
			name: "Ошибка репозитория при обновлении флага заморозки",
			mockBehavior: func(m *mocks.MockClientProfileRepository) {
				m.EXPECT().GetByAccountID(gomock.Any(), accountID).Return(domain.ClientProfile{AccountID: accountID}, nil)
				m.EXPECT().UpdateStreakFreeze(gomock.Any(), accountID, true).Return(errors.New("update error"))
			},
			wantErr: errutil.Internal("failed to activate streak freeze", errors.New("update error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockClientProfileRepository(ctrl)
			tt.mockBehavior(repo)

			uc := NewClientProfileUseCase(repo, nil, nil)
			err := uc.ActivateStreakFreeze(ctx, accountID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClientProfileUseCase_IncrementStreak(t *testing.T) {
	ctx := context.Background()
	accountID := int64(15)
	now := time.Now()
	twoWeeksAgo := now.AddDate(0, 0, -14)

	tests := []struct {
		name         string
		mockBehavior func(m *mocks.MockClientProfileRepository)
		wantErr      error
	}{
		{
			name: "Успешное увеличение стрика без синхронизации",
			mockBehavior: func(m *mocks.MockClientProfileRepository) {
				m.EXPECT().GetByAccountID(gomock.Any(), accountID).Return(domain.ClientProfile{AccountID: accountID}, nil)
				m.EXPECT().IncrementStreak(gomock.Any(), accountID).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "Ошибка получения профиля",
			mockBehavior: func(m *mocks.MockClientProfileRepository) {
				m.EXPECT().GetByAccountID(gomock.Any(), accountID).Return(domain.ClientProfile{}, errors.New("get error"))
			},
			wantErr: errors.New("get error"),
		},
		{
			name: "Ошибка при синхронизации стрика перед инкрементом",
			mockBehavior: func(m *mocks.MockClientProfileRepository) {
				m.EXPECT().GetByAccountID(gomock.Any(), accountID).Return(domain.ClientProfile{
					AccountID:     accountID,
					LastOrderDate: &twoWeeksAgo,
				}, nil)
				m.EXPECT().ResetStreak(gomock.Any(), accountID).Return(errors.New("reset sync error"))
			},
			wantErr: errors.New("reset sync error"),
		},
		{
			name: "Ошибка репозитория при инкременте",
			mockBehavior: func(m *mocks.MockClientProfileRepository) {
				m.EXPECT().GetByAccountID(gomock.Any(), accountID).Return(domain.ClientProfile{AccountID: accountID}, nil)
				m.EXPECT().IncrementStreak(gomock.Any(), accountID).Return(errors.New("increment error"))
			},
			wantErr: errutil.Internal("failed to increment streak", errors.New("increment error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockClientProfileRepository(ctrl)
			tt.mockBehavior(repo)

			uc := NewClientProfileUseCase(repo, nil, nil)
			err := uc.IncrementStreak(ctx, accountID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClientProfileUseCase_ClaimWheelSpin(t *testing.T) {
	ctx := context.Background()
	accountID := int64(99)

	tests := []struct {
		name         string
		mockBehavior func(m *mocks.MockClientProfileRepository)
		wantErr      error
	}{
		{
			name: "Успешное использование прокрутки колеса",
			mockBehavior: func(m *mocks.MockClientProfileRepository) {
				m.EXPECT().ClaimWheelSpin(gomock.Any(), accountID).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "Ошибка базы данных при использовании прокрутки",
			mockBehavior: func(m *mocks.MockClientProfileRepository) {
				m.EXPECT().ClaimWheelSpin(gomock.Any(), accountID).Return(errors.New("claim error"))
			},
			wantErr: errors.New("claim error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockClientProfileRepository(ctrl)
			tt.mockBehavior(repo)

			uc := NewClientProfileUseCase(repo, nil, nil)
			err := uc.ClaimWheelSpin(ctx, accountID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClientProfileUseCase_ResetWheelSpinCooldown(t *testing.T) {
	ctx := context.Background()
	accountID := int64(100)

	tests := []struct {
		name         string
		mockBehavior func(m *mocks.MockClientProfileRepository)
		wantErr      error
	}{
		{
			name: "Успешный сброс кулдауна прокрутки колеса",
			mockBehavior: func(m *mocks.MockClientProfileRepository) {
				m.EXPECT().ResetWheelSpinCooldown(gomock.Any(), accountID).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "Ошибка базы данных при сбросе кулдауна",
			mockBehavior: func(m *mocks.MockClientProfileRepository) {
				m.EXPECT().ResetWheelSpinCooldown(gomock.Any(), accountID).Return(errors.New("cooldown error"))
			},
			wantErr: errors.New("cooldown error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockClientProfileRepository(ctrl)
			tt.mockBehavior(repo)

			uc := NewClientProfileUseCase(repo, nil, nil)
			err := uc.ResetWheelSpinCooldown(ctx, accountID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
