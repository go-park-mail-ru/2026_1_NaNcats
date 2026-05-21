package analyticsclient

//go:generate mockgen -destination=mocks/analytics_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/analyticsclient AnalyticsClient

import (
	"context"
	"errors"
	"time"

	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/analytics"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ErrNotFound        = errors.New("data not found")
	ErrInvalidArgument = errors.New("invalid request parameters")
	ErrForbidden       = errors.New("access denied")
	ErrInternal        = errors.New("internal server error")
)

type FinancialStats struct {
	TotalRevenueRaw   int64
	AverageTicketRaw  int64
	TotalDiscountsRaw int64
	TotalOrdersCount  int64
}

type OperationalStats struct {
	AvgCookingTimeSec int64
	StatusCounts      map[string]int64
}

type BestSeller struct {
	DishID          int64
	DishName        string
	UnitsSold       int32
	TotalRevenueRaw int64
}

type OrderTypeStat struct {
	OrderType    string
	OrdersCount  int64
	AvgGroupSize float64
}

type DailyStat struct {
	Date        time.Time
	RevenueRaw  int64
	OrdersCount int64
}

type OwnerStats struct {
	Financial   FinancialStats
	Operational OperationalStats
	Dishes      []BestSeller
	OrderTypes  []OrderTypeStat
	Timeline    []DailyStat
}

type AnalyticsClient interface {
	// Запрашивает аналитику по ресторану за выбранный период из analytics_service
	GetOwnerStats(ctx context.Context, restaurantID int64, startTime, endTime time.Time) (OwnerStats, error)
}

type analyticsClient struct {
	client pb.AnalyticsServiceClient
}

func NewAnalyticsClient(cl pb.AnalyticsServiceClient) AnalyticsClient {
	return &analyticsClient{
		client: cl,
	}
}

func (c *analyticsClient) GetOwnerStats(ctx context.Context, restaurantID int64, startTime, endTime time.Time) (OwnerStats, error) {
	resp, err := c.client.GetOwnerStats(ctx, &pb.GetOwnerStatsRequest{
		RestaurantId: restaurantID,
		StartTime:    timestamppb.New(startTime),
		EndTime:      timestamppb.New(endTime),
	})
	if err != nil {
		return OwnerStats{}, mapError(err)
	}

	return mapFromPBResponse(resp), nil
}

func mapFromPBResponse(resp *pb.GetOwnerStatsResponse) OwnerStats {
	if resp == nil {
		return OwnerStats{}
	}

	stats := OwnerStats{
		Financial: FinancialStats{
			TotalRevenueRaw:   resp.Financial.TotalRevenueRaw,
			AverageTicketRaw:  resp.Financial.AverageTicketRaw,
			TotalDiscountsRaw: resp.Financial.TotalDiscountsRaw,
			TotalOrdersCount:  resp.Financial.TotalOrdersCount,
		},
		Operational: OperationalStats{
			AvgCookingTimeSec: resp.Operational.AvgCookingTimeSec,
			StatusCounts:      resp.Operational.StatusCounts,
		},
		Dishes:     make([]BestSeller, 0, len(resp.Dishes)),
		OrderTypes: make([]OrderTypeStat, 0, len(resp.OrderTypes)),
		Timeline:   make([]DailyStat, 0, len(resp.Timeline)),
	}

	for _, d := range resp.Dishes {
		stats.Dishes = append(stats.Dishes, BestSeller{
			DishID:          d.DishId,
			DishName:        d.DishName,
			UnitsSold:       d.UnitsSold,
			TotalRevenueRaw: d.TotalRevenueRaw,
		})
	}

	for _, s := range resp.OrderTypes {
		stats.OrderTypes = append(stats.OrderTypes, OrderTypeStat{
			OrderType:    s.OrderType,
			OrdersCount:  s.OrdersCount,
			AvgGroupSize: s.AvgGroupSize,
		})
	}

	for _, t := range resp.Timeline {
		var date time.Time
		if t.Date != nil {
			date = t.Date.AsTime()
		}
		stats.Timeline = append(stats.Timeline, DailyStat{
			Date:        date,
			RevenueRaw:  t.RevenueRaw,
			OrdersCount: t.OrdersCount,
		})
	}

	return stats
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	st, ok := status.FromError(err)
	if !ok {
		return ErrInternal
	}
	switch st.Code() {
	case codes.NotFound:
		return ErrNotFound
	case codes.InvalidArgument:
		return ErrInvalidArgument
	case codes.PermissionDenied:
		return ErrForbidden
	default:
		return ErrInternal
	}
}
