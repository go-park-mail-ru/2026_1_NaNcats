package rabbitmq

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"
	"github.com/mailru/easyjson"
)

type PaymentConsumer struct {
	client  *rabbitmq.RabbitClient
	usecase usecase.PaymentUseCase
	logger  logger.Logger
}

func NewPaymentConsumer(client *rabbitmq.RabbitClient, uc usecase.PaymentUseCase, l logger.Logger) *PaymentConsumer {
	return &PaymentConsumer{
		client:  client,
		usecase: uc,
		logger:  l,
	}
}

func (c *PaymentConsumer) Start(ctx context.Context) error {
	handler := func(body []byte) error {
		var cmd events.SagaCommand
		if err := easyjson.Unmarshal(body, &cmd); err != nil {
			c.logger.Error("failed to unmarshal payment saga command", err)
			return nil
		}

		c.logger.Info("Received payment command", logger.String("order_id", cmd.OrderID))

		reply := events.SagaReply{
			OrderID: cmd.OrderID,
			Step:    "PAYMENT",
		}

		if cmd.Action == events.CommandCreatePayment {
			paymentID, confirmationURL, err := c.usecase.CreatePayment(ctx, cmd.Amount, cmd.PaymentMethodID, cmd.IdempotencyKey)

			if err != nil {
				reply.Status = events.StatusError
				reply.ErrorMessage = err.Error()
				c.logger.Error("Payment creation failed", err)
			} else {
				reply.Status = events.StatusSuccess
				reply.PaymentID = paymentID
				reply.PaymentURL = confirmationURL
			}
		}

		err := c.client.PublishJSON(ctx, events.QueueOrderReplies, reply)
		if err != nil {
			c.logger.Error("failed to publish payment reply", err)
			return err
		}

		return nil
	}

	return c.client.ConsumeJSON(ctx, events.QueuePaymentCommands, handler)
}
