package rabbitmq

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"

	"github.com/mailru/easyjson"
)

type OrderConsumer struct {
	client  *rabbitmq.RabbitClient
	usecase usecase.OrderUseCase
	logger  logger.Logger
}

func NewOrderConsumer(client *rabbitmq.RabbitClient, uc usecase.OrderUseCase, l logger.Logger) *OrderConsumer {
	return &OrderConsumer{
		client:  client,
		usecase: uc,
		logger:  l,
	}
}

func (c *OrderConsumer) Start(ctx context.Context) error {
	handler := func(ctx context.Context, body []byte) error {
		var reply events.SagaReply

		if err := easyjson.Unmarshal(body, &reply); err != nil {
			c.logger.Error("failed to unmarshal saga reply", err)
			return nil
		}

		c.logger.Info("Received saga reply", logger.Any("reply", reply))

		return c.usecase.ProcessSagaReply(ctx, reply)
	}

	return c.client.ConsumeJSON(ctx, events.QueueOrderReplies, handler)
}
