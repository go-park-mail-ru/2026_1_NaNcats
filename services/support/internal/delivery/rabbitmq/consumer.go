package rabbitmq

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/support/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"
	"github.com/mailru/easyjson"
)

type SupportConsumer struct {
	rabbit *rabbitmq.RabbitClient
	repo   repository.SupportRepository
	logger logger.Logger
}

func NewSupportConsumer(r *rabbitmq.RabbitClient, repo repository.SupportRepository, l logger.Logger) *SupportConsumer {
	return &SupportConsumer{
		rabbit: r,
		repo:   repo,
		logger: l,
	}
}

func (c *SupportConsumer) Start(ctx context.Context) error {
	// Определяем обработчик входящих сообщений
	handler := func(ctx context.Context, body []byte) error {
		var event events.UserRoleChangedEvent
		if err := easyjson.Unmarshal(body, &event); err != nil {
			c.logger.Error("failed to unmarshal user event", err)
			return nil // Возвращаем nil, чтобы не зацикливать битое сообщение в очереди
		}

		c.logger.Info("Received UserRoleChanged event",
			logger.Int64("user_id", event.UserID),
			logger.String("new_role", event.NewRole),
		)

		// Если пользователь стал саппортом - создаем ему профиль агента
		if event.NewRole == "support" {
			err := c.repo.CreateAgentProfile(ctx, event.UserID)
			if err != nil {
				c.logger.Error("failed to create agent profile", err)
				return err // Возвращаем ошибку, RabbitMQ переотправит сообщение позже
			}
			c.logger.Info("Agent profile created successfully", logger.Int64("user_id", event.UserID))
		}

		// Если пользователь перестал быть саппортом
		if event.OldRole == "support" && event.NewRole != "support" {
			err := c.repo.DeleteAgentProfile(ctx, event.UserID)
			if err != nil {
				c.logger.Error("failed to delete agent profile", err)
				return err
			}
			c.logger.Info("Agent profile removed", logger.Int64("user_id", event.UserID))
		}

		return nil
	}

	// Запускаем прослушивание очереди ивентов юзера
	return c.rabbit.ConsumeJSON(ctx, events.QueueUserEvents, handler)
}
