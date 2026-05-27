package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/repository"
	ucMocks "github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/usecase/mocks"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/analytics"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestAnalyticsHandler_GetOwnerStats(t *testing.T) {
	type mockInit func(uc *ucMocks.MockAnalyticsUseCase)

	startT := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	endT := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	fullStats := repository.OwnerStats{
		Financial: repository.FinancialStats{
			TotalRevenueRaw:   500_000_000,
			AverageTicketRaw:  100_000_000,
			TotalDiscountsRaw: 10_000_000,
			TotalOrdersCount:  5,
		},
		Operational: repository.OperationalStats{
			AvgCookingTimeSec: 720,
			StatusCounts:      map[string]int64{"paid": 5, "finished": 4, "cancelled": 1},
		},
		Dishes: []repository.BestSeller{
			{DishID: 7, DishName: "Бургер", UnitsSold: 12, TotalRevenueRaw: 240_000_000},
			{DishID: 8, DishName: "Картошка", UnitsSold: 9, TotalRevenueRaw: 90_000_000},
		},
		OrderTypes: []repository.OrderTypeStat{
			{OrderType: "solo", OrdersCount: 3, AvgGroupSize: 1.0},
			{OrderType: "shared", OrdersCount: 2, AvgGroupSize: 3.5},
		},
		Timeline: []repository.DailyStat{
			{Date: startT, RevenueRaw: 200_000_000, OrdersCount: 2},
			{Date: endT, RevenueRaw: 300_000_000, OrdersCount: 3},
		},
	}

	tests := []struct {
		name          string
		req           *pb.GetOwnerStatsRequest
		mockInit      mockInit
		expectedCode  codes.Code
		assertSuccess func(t *testing.T, resp *pb.GetOwnerStatsResponse)
	}{
		{
			name: "Успешное получение полной статистики",
			req: &pb.GetOwnerStatsRequest{
				RestaurantId: 100,
				StartTime:    timestamppb.New(startT),
				EndTime:      timestamppb.New(endT),
			},
			mockInit: func(uc *ucMocks.MockAnalyticsUseCase) {
				uc.EXPECT().
					GetOwnerStats(gomock.Any(), int64(100), gomock.Any(), gomock.Any()).
					Return(fullStats, nil)
			},
			expectedCode: codes.OK,
			assertSuccess: func(t *testing.T, resp *pb.GetOwnerStatsResponse) {
				assert.EqualValues(t, 500_000_000, resp.Financial.TotalRevenueRaw)
				assert.EqualValues(t, 5, resp.Financial.TotalOrdersCount)
				assert.EqualValues(t, 720, resp.Operational.AvgCookingTimeSec)
				assert.Equal(t, int64(5), resp.Operational.StatusCounts["paid"])
				assert.Len(t, resp.Dishes, 2)
				assert.Equal(t, "Бургер", resp.Dishes[0].DishName)
				assert.Len(t, resp.OrderTypes, 2)
				assert.Equal(t, "shared", resp.OrderTypes[1].OrderType)
				assert.Len(t, resp.Timeline, 2)
				assert.True(t, resp.Timeline[0].Date.AsTime().Equal(startT))
			},
		},
		{
			name: "Отсутствует start_time — InvalidArgument",
			req: &pb.GetOwnerStatsRequest{
				RestaurantId: 100,
				StartTime:    nil,
				EndTime:      timestamppb.New(endT),
			},
			mockInit:     func(uc *ucMocks.MockAnalyticsUseCase) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Отсутствует end_time — InvalidArgument",
			req: &pb.GetOwnerStatsRequest{
				RestaurantId: 100,
				StartTime:    timestamppb.New(startT),
				EndTime:      nil,
			},
			mockInit:     func(uc *ucMocks.MockAnalyticsUseCase) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Чужой ресторан — PermissionDenied",
			req: &pb.GetOwnerStatsRequest{
				RestaurantId: 999,
				StartTime:    timestamppb.New(startT),
				EndTime:      timestamppb.New(endT),
			},
			mockInit: func(uc *ucMocks.MockAnalyticsUseCase) {
				uc.EXPECT().GetOwnerStats(gomock.Any(), int64(999), gomock.Any(), gomock.Any()).
					Return(repository.OwnerStats{}, domain.ErrPermissionDenied)
			},
			expectedCode: codes.PermissionDenied,
		},
		{
			name: "Анонимный запрос — Unauthenticated",
			req: &pb.GetOwnerStatsRequest{
				RestaurantId: 100,
				StartTime:    timestamppb.New(startT),
				EndTime:      timestamppb.New(endT),
			},
			mockInit: func(uc *ucMocks.MockAnalyticsUseCase) {
				uc.EXPECT().GetOwnerStats(gomock.Any(), int64(100), gomock.Any(), gomock.Any()).
					Return(repository.OwnerStats{}, domain.ErrUnauthorized)
			},
			expectedCode: codes.Unauthenticated,
		},
		{
			name: "Ресторан не найден — NotFound",
			req: &pb.GetOwnerStatsRequest{
				RestaurantId: 100,
				StartTime:    timestamppb.New(startT),
				EndTime:      timestamppb.New(endT),
			},
			mockInit: func(uc *ucMocks.MockAnalyticsUseCase) {
				uc.EXPECT().GetOwnerStats(gomock.Any(), int64(100), gomock.Any(), gomock.Any()).
					Return(repository.OwnerStats{}, domain.ErrRestaurantNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Произвольная ошибка usecase — Internal",
			req: &pb.GetOwnerStatsRequest{
				RestaurantId: 100,
				StartTime:    timestamppb.New(startT),
				EndTime:      timestamppb.New(endT),
			},
			mockInit: func(uc *ucMocks.MockAnalyticsUseCase) {
				uc.EXPECT().GetOwnerStats(gomock.Any(), int64(100), gomock.Any(), gomock.Any()).
					Return(repository.OwnerStats{}, errors.New("unexpected"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			uc := ucMocks.NewMockAnalyticsUseCase(ctrl)
			tt.mockInit(uc)

			h := NewAnalyticsHandler(uc)
			resp, err := h.GetOwnerStats(context.Background(), tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.assertSuccess != nil {
					tt.assertSuccess(t, resp)
				}
			} else {
				st, ok := status.FromError(err)
				assert.True(t, ok, "expected grpc status error")
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestAnalyticsHandler_GetOwnerStats_EmptyMappingsReturnEmptySlices(t *testing.T) {
	// Проверяем, что пустой ответ репозитория не превращается в nil-слайсы:
	// gRPC по умолчанию сериализует nil-слайсы как nil, мы хотим явные пустые массивы.
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := ucMocks.NewMockAnalyticsUseCase(ctrl)
	uc.EXPECT().GetOwnerStats(gomock.Any(), int64(1), gomock.Any(), gomock.Any()).
		Return(repository.OwnerStats{}, nil)

	h := NewAnalyticsHandler(uc)
	resp, err := h.GetOwnerStats(context.Background(), &pb.GetOwnerStatsRequest{
		RestaurantId: 1,
		StartTime:    timestamppb.New(time.Now().Add(-time.Hour)),
		EndTime:      timestamppb.New(time.Now()),
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Dishes)
	assert.NotNil(t, resp.OrderTypes)
	assert.NotNil(t, resp.Timeline)
	assert.Empty(t, resp.Dishes)
	assert.Empty(t, resp.OrderTypes)
	assert.Empty(t, resp.Timeline)
}
