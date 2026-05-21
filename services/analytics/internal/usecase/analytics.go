package usecase

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
)

// Интерфейс для обработки входящих аналитических событий
type AnalyticsUseCase interface {
	ProcessEvent(ctx context.Context, event events.AnalyticsOrderEvent) error
	GetOwnerStats(ctx context.Context, restaurantID int64, startTime, endTime time.Time) (repository.OwnerStats, error)
}

type analyticsUseCase struct {
	repo   repository.AnalyticsRepository
	logger logger.Logger
}

func NewAnalyticsUseCase(repo repository.AnalyticsRepository, l logger.Logger) AnalyticsUseCase {
	return &analyticsUseCase{
		repo:   repo,
		logger: l,
	}
}

// Выполняет бизнес-валидацию и сохраняет событие в базу
func (u *analyticsUseCase) ProcessEvent(ctx context.Context, event events.AnalyticsOrderEvent) error {
	if event.EventTime == 0 {
		event.EventTime = time.Now().UnixMilli()
	}

	if event.Status == "paid" {
		event.IsFinancialImpact = 1
	} else {
		event.IsFinancialImpact = 0
	}

	if err := u.repo.InsertEvent(ctx, event); err != nil {
		u.logger.Error("failed to process analytics event", err,
			logger.String("order_id", event.OrderPublicID),
			logger.String("status", event.Status),
		)
		return err // Возвращаем наверх, чтобы RabbitMQ сделал retry
	}

	return nil
}

func (u *analyticsUseCase) GetOwnerStats(ctx context.Context, restaurantID int64, startTime, endTime time.Time) (repository.OwnerStats, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("restaurant.id", restaurantID),
		attribute.String("query.start_time", startTime.Format(time.RFC3339)),
		attribute.String("query.end_time", endTime.Format(time.RFC3339)),
	)

	if startTime.After(endTime) {
		return repository.OwnerStats{}, errutil.New("INVALID_TIME_RANGE", "start_time must be before end_time", codes.InvalidArgument)
	}

	stats, err := u.repo.GetOwnerStats(ctx, restaurantID, startTime, endTime)
	if err != nil {
		u.logger.Error("failed to fetch owner stats from repository", err,
			logger.Int64("restaurant_id", restaurantID),
		)
		return repository.OwnerStats{}, errutil.Internal("failed to fetch analytics data", err)
	}

	return stats, nil
}
