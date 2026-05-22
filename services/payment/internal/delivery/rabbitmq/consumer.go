package rabbitmq

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"
	"github.com/mailru/easyjson"
)

const (
	// Период и предел опроса статуса платежа в YooKassa.
	paymentPollInterval = 5 * time.Second
	paymentPollAttempts = 120
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

// pollPaymentUntilSettled подстраховывает доставку статуса платежа на случай,
// когда вебхук YooKassa не доходит до сервиса (типично для локальной
// разработки: YooKassa не достучится до localhost). Периодически тянет статус
// платежа из YooKassa REST и применяет его тем же путём, что и вебхук;
// останавливается, как только платёж в терминальном статусе либо по таймауту.
func (c *PaymentConsumer) pollPaymentUntilSettled(ctx context.Context, paymentID string) {
	if paymentID == "" {
		return
	}
	for attempt := 0; attempt < paymentPollAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return
		case <-time.After(paymentPollInterval):
		}

		status, err := c.usecase.RefreshPaymentStatus(ctx, paymentID)
		if err != nil {
			continue
		}
		if status == "succeeded" || status == "canceled" {
			c.logger.Info("payment poller: payment settled",
				logger.String("payment_id", paymentID), logger.String("status", status))
			return
		}
	}
}

func (c *PaymentConsumer) Start(ctx context.Context) error {
	svcCtx := ctx
	handler := func(ctx context.Context, body []byte) error {
		var cmd events.SagaCommand
		if err := easyjson.Unmarshal(body, &cmd); err != nil {
			c.logger.Error("failed to unmarshal payment saga command", err)
			return nil
		}

		c.logger.Info("Received payment command",
			logger.String("order_id", cmd.OrderID),
			logger.String("split_id", cmd.SplitID),
			logger.Int64("amount", cmd.Amount),
			logger.String("payment_method_id", cmd.PaymentMethodID),
			logger.String("action", cmd.Action),
		)

		reply := events.SagaReply{
			OrderID: cmd.OrderID,
			SplitID: cmd.SplitID,
			UserID:  cmd.UserID,
			Step:    "PAYMENT",
		}

		if cmd.Action == events.CommandCreatePayment {
			paymentID, confirmationURL, err := c.usecase.CreatePayment(ctx, cmd.Amount, cmd.PaymentMethodID, cmd.IdempotencyKey)

			if err != nil {
				reply.Status = events.StatusError
				reply.ErrorMessage = err.Error()
				c.logger.Error("Payment creation failed", err,
					logger.String("order_id", cmd.OrderID),
					logger.String("split_id", cmd.SplitID),
					logger.Int64("amount", cmd.Amount),
					logger.String("payment_method_id", cmd.PaymentMethodID),
				)
			} else {
				reply.Status = events.StatusSuccess
				reply.PaymentID = paymentID
				reply.PaymentURL = confirmationURL
				// Подстраховка на случай недоставленного вебхука YooKassa.
				go c.pollPaymentUntilSettled(svcCtx, paymentID)
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
