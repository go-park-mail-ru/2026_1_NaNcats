package rabbitmq

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/delivery/websocket"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"
	"github.com/mailru/easyjson"
)

type GatewayConsumer struct {
	client    *rabbitmq.RabbitClient
	wsManager *websocket.WsManager
	logger    logger.Logger
}

func NewGatewayConsumer(client *rabbitmq.RabbitClient, wsManager *websocket.WsManager, l logger.Logger) *GatewayConsumer {
	return &GatewayConsumer{
		client:    client,
		wsManager: wsManager,
		logger:    l,
	}
}

func (c *GatewayConsumer) Start(ctx context.Context) error {
	handler := func(ctx context.Context, body []byte) error {
		var event events.GatewayEvent
		if err := easyjson.Unmarshal(body, &event); err != nil {
			c.logger.Error("failed to unmarshal gateway event", err)
			return nil
		}

		if event.OrderID != "" {
			c.logger.Info("Received Order event", logger.String("order_id", event.OrderID))
		} else if event.CartID != "" {
			c.logger.Info("Received Cart event", logger.String("cart_id", event.CartID), logger.String("event_type", event.EventType))
		}

		return c.wsManager.BroadcastToRedis(event)
	}

	return c.client.ConsumeJSON(ctx, events.QueueGatewayEvents, handler)
}
