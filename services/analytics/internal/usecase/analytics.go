package usecase

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"
)

// Интерфейс для обработки входящих аналитических событий
type AnalyticsUseCase interface {
	ProcessEvent(ctx context.Context, event events.AnalyticsOrderEvent) error
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
