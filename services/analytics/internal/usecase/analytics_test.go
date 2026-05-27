package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/repository"
	repoMocks "github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/repository/mocks"
	ucMocks "github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/usecase/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAnalyticsUseCase_ProcessEvent(t *testing.T) {
	type mockInit func(r *repoMocks.MockAnalyticsRepository)

	tests := []struct {
		name           string
		input          events.AnalyticsOrderEvent
		mockInit       mockInit
		expectErr      bool
		assertInserted func(t *testing.T, got events.AnalyticsOrderEvent)
	}{
		{
			name: "Статус paid выставляет is_financial_impact=1",
			input: events.AnalyticsOrderEvent{
				OrderPublicID: "ord-1",
				Status:        "paid",
				EventTime:     1700000000000,
			},
			mockInit: func(r *repoMocks.MockAnalyticsRepository) {
				r.EXPECT().InsertEvent(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, e events.AnalyticsOrderEvent) error {
						assert.EqualValues(t, 1, e.IsFinancialImpact)
						return nil
					})
			},
		},
		{
			name: "Не-paid статус не помечается финансовым",
			input: events.AnalyticsOrderEvent{
				OrderPublicID: "ord-2",
				Status:        "finished",
				EventTime:     1700000000000,
			},
			mockInit: func(r *repoMocks.MockAnalyticsRepository) {
				r.EXPECT().InsertEvent(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, e events.AnalyticsOrderEvent) error {
						assert.EqualValues(t, 0, e.IsFinancialImpact)
						return nil
					})
			},
		},
		{
			name: "Пустой EventTime подставляется текущим временем",
			input: events.AnalyticsOrderEvent{
				OrderPublicID: "ord-3",
				Status:        "in_progress",
				EventTime:     0,
			},
			mockInit: func(r *repoMocks.MockAnalyticsRepository) {
				r.EXPECT().InsertEvent(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, e events.AnalyticsOrderEvent) error {
						assert.NotZero(t, e.EventTime, "event_time должен быть автозаполнен")
						return nil
					})
			},
		},
		{
			name: "Ошибка репозитория пробрасывается наверх (для RabbitMQ retry)",
			input: events.AnalyticsOrderEvent{
				OrderPublicID: "ord-4",
				Status:        "paid",
				EventTime:     1700000000000,
			},
			mockInit: func(r *repoMocks.MockAnalyticsRepository) {
				r.EXPECT().InsertEvent(gomock.Any(), gomock.Any()).
					Return(errors.New("clickhouse unavailable"))
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := repoMocks.NewMockAnalyticsRepository(ctrl)
			tt.mockInit(repo)

			uc := NewAnalyticsUseCase(repo, logger.NewNopLogger(), nil)

			err := uc.ProcessEvent(context.Background(), tt.input)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAnalyticsUseCase_GetOwnerStats(t *testing.T) {
	const ownerID int64 = 42
	const restaurantID int64 = 100

	ctxWithOwner := context.WithValue(context.Background(), common.UserIDKey, ownerID)
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	startTime := now.AddDate(0, 0, -7)
	endTime := now

	type mockInit func(r *repoMocks.MockAnalyticsRepository, rc *ucMocks.MockRestaurantClient)

	expectedStats := repository.OwnerStats{
		Financial: repository.FinancialStats{
			TotalRevenueRaw:  100_000_000,
			AverageTicketRaw: 25_000_000,
			TotalOrdersCount: 4,
		},
		Operational: repository.OperationalStats{AvgCookingTimeSec: 600},
	}

	tests := []struct {
		name        string
		ctx         context.Context
		startTime   time.Time
		endTime     time.Time
		mockInit    mockInit
		expectStats repository.OwnerStats
		expectErr   error
	}{
		{
			name:      "Успешное получение статистики владельцем",
			ctx:       ctxWithOwner,
			startTime: startTime,
			endTime:   endTime,
			mockInit: func(r *repoMocks.MockAnalyticsRepository, rc *ucMocks.MockRestaurantClient) {
				rc.EXPECT().GetRestaurantBrandByID(gomock.Any(), restaurantID).
					Return(domain.RestaurantBrand{ID: restaurantID, OwnerProfileID: ownerID, Name: "Burger Heroes"}, nil)
				r.EXPECT().GetOwnerStats(gomock.Any(), restaurantID, startTime, endTime).
					Return(expectedStats, nil)
			},
			expectStats: expectedStats,
		},
		{
			name:      "Запрос без user_id в контексте — Unauthorized",
			ctx:       context.Background(),
			startTime: startTime,
			endTime:   endTime,
			mockInit:  func(r *repoMocks.MockAnalyticsRepository, rc *ucMocks.MockRestaurantClient) {},
			expectErr: domain.ErrUnauthorized,
		},
		{
			name:      "startTime > endTime — InvalidArgument",
			ctx:       ctxWithOwner,
			startTime: endTime,
			endTime:   startTime,
			mockInit:  func(r *repoMocks.MockAnalyticsRepository, rc *ucMocks.MockRestaurantClient) {},
			expectErr: errors.New("start_time must be before end_time"),
		},
		{
			name:      "Ресторан не найден — ошибка пробрасывается",
			ctx:       ctxWithOwner,
			startTime: startTime,
			endTime:   endTime,
			mockInit: func(r *repoMocks.MockAnalyticsRepository, rc *ucMocks.MockRestaurantClient) {
				rc.EXPECT().GetRestaurantBrandByID(gomock.Any(), restaurantID).
					Return(domain.RestaurantBrand{}, domain.ErrRestaurantNotFound)
			},
			expectErr: domain.ErrRestaurantNotFound,
		},
		{
			name:      "Ресторан принадлежит другому owner — PermissionDenied",
			ctx:       ctxWithOwner,
			startTime: startTime,
			endTime:   endTime,
			mockInit: func(r *repoMocks.MockAnalyticsRepository, rc *ucMocks.MockRestaurantClient) {
				rc.EXPECT().GetRestaurantBrandByID(gomock.Any(), restaurantID).
					Return(domain.RestaurantBrand{ID: restaurantID, OwnerProfileID: ownerID + 1, Name: "Чужой"}, nil)
			},
			expectErr: domain.ErrPermissionDenied,
		},
		{
			name:      "Ошибка репозитория — Internal",
			ctx:       ctxWithOwner,
			startTime: startTime,
			endTime:   endTime,
			mockInit: func(r *repoMocks.MockAnalyticsRepository, rc *ucMocks.MockRestaurantClient) {
				rc.EXPECT().GetRestaurantBrandByID(gomock.Any(), restaurantID).
					Return(domain.RestaurantBrand{ID: restaurantID, OwnerProfileID: ownerID}, nil)
				r.EXPECT().GetOwnerStats(gomock.Any(), restaurantID, startTime, endTime).
					Return(repository.OwnerStats{}, errors.New("clickhouse error"))
			},
			expectErr: errors.New("failed to fetch analytics data"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := repoMocks.NewMockAnalyticsRepository(ctrl)
			restClient := ucMocks.NewMockRestaurantClient(ctrl)
			tt.mockInit(repo, restClient)

			uc := NewAnalyticsUseCase(repo, logger.NewNopLogger(), restClient)
			stats, err := uc.GetOwnerStats(tt.ctx, restaurantID, tt.startTime, tt.endTime)

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectStats, stats)
			}
		})
	}
}
