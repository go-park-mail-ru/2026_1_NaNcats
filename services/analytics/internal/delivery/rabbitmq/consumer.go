package rabbitmq

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq"
	rabbitmqErrors "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/errors"
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
	// Запускаем прослушивание очереди аналитики
	return c.rabbit.ConsumeJSON(ctx, events.QueueAnalytics, c.handleMessage)
}

// handleMessage обрабатывает одно сообщение из очереди аналитики.
// Невалидный JSON отбрасывается как permanent error — RabbitMQ не вернёт
// такое сообщение в очередь. Ошибка от usecase (например, ClickHouse недоступен)
// пробрасывается наверх, чтобы брокер сделал retry.
func (c *AnalyticsConsumer) handleMessage(ctx context.Context, body []byte) error {
	var event events.AnalyticsOrderEvent
	if err := easyjson.Unmarshal(body, &event); err != nil {
		c.logger.Error("failed to unmarshal analytics event, message dropped", err)
		return rabbitmqErrors.NewPermanentError(fmt.Errorf("unmarshal payload failed: %w", err))
	}

	c.logger.Debug("received analytics event",
		logger.String("order_id", event.OrderPublicID),
		logger.String("status", event.Status),
	)

	if err := c.uc.ProcessEvent(ctx, event); err != nil {
		return err
	}

	return nil
}
