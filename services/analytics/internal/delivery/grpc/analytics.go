package grpc

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/grpcutil"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/analytics"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AnalyticsHandler struct {
	pb.UnimplementedAnalyticsServiceServer
	usecase usecase.AnalyticsUseCase
}

func NewAnalyticsHandler(uc usecase.AnalyticsUseCase) *AnalyticsHandler {
	return &AnalyticsHandler{
		usecase: uc,
	}
}

func (h *AnalyticsHandler) GetOwnerStats(ctx context.Context, req *pb.GetOwnerStatsRequest) (*pb.GetOwnerStatsResponse, error) {
	// Обязательные поля времени
	if req.StartTime == nil || req.EndTime == nil {
		return nil, status.Error(codes.InvalidArgument, "start_time and end_time are required parameters")
	}

	// Конвертируем gRPC Timestamp в time.Time
	startTime := req.StartTime.AsTime()
	endTime := req.EndTime.AsTime()

	stats, err := h.usecase.GetOwnerStats(ctx, req.RestaurantId, startTime, endTime)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &pb.GetOwnerStatsResponse{
		Financial:   mapDomainToPBFinancial(stats.Financial),
		Operational: mapDomainToPBOperational(stats.Operational),
		Dishes:      mapDomainToPBBestSellers(stats.Dishes),
		OrderTypes:  mapDomainToPBOrderTypes(stats.OrderTypes),
		Timeline:    mapDomainToPBTimeline(stats.Timeline),
	}, nil
}

func mapDomainToPBFinancial(f repository.FinancialStats) *pb.FinancialStats {
	return &pb.FinancialStats{
		TotalRevenueRaw:   f.TotalRevenueRaw,
		AverageTicketRaw:  f.AverageTicketRaw,
		TotalDiscountsRaw: f.TotalDiscountsRaw,
		TotalOrdersCount:  f.TotalOrdersCount,
	}
}

func mapDomainToPBOperational(o repository.OperationalStats) *pb.OperationalStats {
	return &pb.OperationalStats{
		AvgCookingTimeSec: o.AvgCookingTimeSec,
		StatusCounts:      o.StatusCounts,
	}
}

func mapDomainToPBBestSellers(dishes []repository.BestSeller) []*pb.BestSeller {
	res := make([]*pb.BestSeller, len(dishes))
	for i, d := range dishes {
		res[i] = &pb.BestSeller{
			DishId:          d.DishID,
			DishName:        d.DishName,
			UnitsSold:       d.UnitsSold,
			TotalRevenueRaw: d.TotalRevenueRaw,
		}
	}
	return res
}

func mapDomainToPBOrderTypes(stats []repository.OrderTypeStat) []*pb.OrderTypeStat {
	res := make([]*pb.OrderTypeStat, len(stats))
	for i, s := range stats {
		res[i] = &pb.OrderTypeStat{
			OrderType:    s.OrderType,
			OrdersCount:  s.OrdersCount,
			AvgGroupSize: s.AvgGroupSize,
		}
	}
	return res
}

func mapDomainToPBTimeline(timeline []repository.DailyStat) []*pb.DailyStat {
	res := make([]*pb.DailyStat, len(timeline))
	for i, p := range timeline {
		res[i] = &pb.DailyStat{
			// Переводим стандартное время в формат gRPC Timestamp
			Date:        timestamppb.New(p.Date),
			RevenueRaw:  p.RevenueRaw,
			OrdersCount: p.OrdersCount,
		}
	}
	return res
}
