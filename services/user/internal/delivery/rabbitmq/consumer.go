package rabbitmq

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq"
	rabbitmqErrors "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/errors"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"
	"github.com/mailru/easyjson"
)

const QueueUserPaid = "user.order_paid"

type UserConsumer struct {
	client        *rabbitmq.RabbitClient
	achievementUC usecase.AchievementUseCase
	logger        logger.Logger
}

func NewUserConsumer(client *rabbitmq.RabbitClient, auc usecase.AchievementUseCase, l logger.Logger) *UserConsumer {
	return &UserConsumer{
		client:        client,
		achievementUC: auc,
		logger:        l,
	}
}

func (c *UserConsumer) Start(ctx context.Context) error {
	handler := func(ctx context.Context, body []byte) error {
		var event events.OrderPaidEvent
		if err := easyjson.Unmarshal(body, &event); err != nil {
			c.logger.Error("user-consumer: failed to unmarshal OrderPaidEvent", err)
			return rabbitmqErrors.NewPermanentError(fmt.Errorf("unmarshal failed: %w", err))
		}

		c.logger.Info("user-consumer: received OrderPaidEvent",
			logger.Int64("user_id", event.UserID),
			logger.Int64("restaurant_id", event.RestaurantID),
			logger.String("order_id", event.OrderPublicID),
		)

		err := c.achievementUC.OnOrderPaid(ctx, event.UserID, event.RestaurantID, event.PaidAt)
		if err != nil {
			c.logger.Error("user-consumer: failed to execute OnOrderPaid UC", err)
			return err
		}

		return nil
	}

	return c.client.ConsumeJSON(ctx, QueueUserPaid, handler)
}
