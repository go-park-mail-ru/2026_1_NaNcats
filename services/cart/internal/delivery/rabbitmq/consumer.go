package rabbitmq

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"

	"github.com/mailru/easyjson"
)

type CartConsumer struct {
	client  *rabbitmq.RabbitClient
	usecase usecase.CartUseCase
	logger  logger.Logger
}

func NewCartConsumer(client *rabbitmq.RabbitClient, uc usecase.CartUseCase, l logger.Logger) *CartConsumer {
	return &CartConsumer{
		client:  client,
		usecase: uc,
		logger:  l,
	}
}

func (c *CartConsumer) Start(ctx context.Context) error {
	handler := func(body []byte) error {
		var cmd events.SagaCommand
		if err := easyjson.Unmarshal(body, &cmd); err != nil {
			c.logger.Error("failed to unmarshal cart saga command", err)
			return nil
		}

		c.logger.Info("Received command in cart service", logger.String("action", cmd.Action))

		var err error
		switch cmd.Action {
		case events.CommandLockCart:
			err = c.usecase.LockCart(ctx, cmd.CartID, cmd.UserID, domain.PaymentIntent{}, cmd.IdempotencyKey)
		case events.CommandUnlockCart:
			err = c.usecase.UnlockCart(ctx, cmd.CartID, cmd.UserID, cmd.IdempotencyKey)
		case events.CommandClearCart:
			err = c.usecase.ClearCart(ctx, cmd.CartID, cmd.UserID, cmd.IdempotencyKey)
		default:
			c.logger.Warn("Unknown action in cart saga", logger.String("action", cmd.Action))
			return nil
		}

		reply := events.SagaReply{
			OrderID: cmd.OrderID,
			Step:    "CART",
			Status:  events.StatusSuccess,
		}

		if err != nil {
			reply.Status = events.StatusError
			reply.ErrorMessage = err.Error()
			c.logger.Error("cart operation failed", err)
		}

		// Пуш ответа в очередь оркестратора
		publishErr := c.client.PublishJSON(ctx, events.QueueOrderReplies, reply)
		if publishErr != nil {
			c.logger.Error("failed to publish reply from cart", publishErr)
			return publishErr
		}

		return nil
	}

	return c.client.ConsumeJSON(ctx, events.QueueCartCommands, handler)
}
