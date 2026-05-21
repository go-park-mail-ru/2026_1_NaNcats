package rabbitmq

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"
	"github.com/mailru/easyjson"
)

type AnalyticsConsumer struct {
	rabbit *rabbitmq.RabbitClient
	uc     usecase.AnalyticsUseCase
	logger logger.Logger
}

func NewAnalyticsConsumer(r *rabbitmq.RabbitClient, uc usecase.AnalyticsUseCase, l logger.Logger) *AnalyticsConsumer {
	return &AnalyticsConsumer{
		rabbit: r,
		uc:     uc,
		logger: l,
	}
}

func (c *AnalyticsConsumer) Start(ctx context.Context) error {
	handler := func(ctx context.Context, body []byte) error {
		var event events.AnalyticsOrderEvent
		if err := easyjson.Unmarshal(body, &event); err != nil {
			c.logger.Error("failed to unmarshal analytics event, message dropped", err)
			return nil // Возвращаем nil, чтобы не зацикливать битый JSON в очереди
		}

		c.logger.Debug("received analytics event",
			logger.String("order_id", event.OrderPublicID),
			logger.String("status", event.Status),
		)

		// Если ClickHouse недоступен, вернется ошибка, и RabbitMQ вернет обратно в очередь
		if err := c.uc.ProcessEvent(ctx, event); err != nil {
			return err
		}

		return nil
	}

	// Запускаем прослушивание очереди аналитики
	return c.rabbit.ConsumeJSON(ctx, events.QueueAnalytics, handler)
}
