package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/mailru/easyjson"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitClient struct {
	conn   *amqp.Connection
	ch     *amqp.Channel
	logger logger.Logger
}

func NewRabbitClient(url string, logger logger.Logger) (*RabbitClient, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	return &RabbitClient{
		conn:   conn,
		ch:     ch,
		logger: logger,
	}, nil
}

func (rc *RabbitClient) Close() {
	if rc.ch != nil {
		rc.ch.Close()
	}
	if rc.conn != nil {
		rc.conn.Close()
	}
}

func (rc *RabbitClient) PublishJSON(ctx context.Context, queueName string, data any) error {
	q, err := rc.ch.QueueDeclare(
		queueName,
		true, false, false, false, nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %w", err)
	}

	var body []byte

	if m, ok := data.(easyjson.Marshaler); ok {
		body, err = easyjson.Marshal(m)
	} else {
		rc.logger.Warn("PublishJSON: structure does not implement easyjson, falling back to slow json", logger.String("type", fmt.Sprintf("%T", data)))
		body, err = json.Marshal(data)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = rc.ch.PublishWithContext(ctx,
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}

	return nil
}

func (rc *RabbitClient) ConsumeJSON(ctx context.Context, queueName string, handler func(body []byte) error) error {
	q, err := rc.ch.QueueDeclare(
		queueName,
		true, false, false, false, nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %w", err)
	}

	// Ограниченное количество сообщений - 10 | защита от перегруза
	err = rc.ch.Qos(10, 0, false)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := rc.ch.ConsumeWithContext(ctx,
		q.Name, "",
		false, false, false, false, nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	go func() {
		for d := range msgs {
			err := handler(d.Body)
			if err != nil {
				rc.logger.Error("failed to process message", err, logger.Field{
					Key:   "queue",
					Value: queueName,
				})
				err := d.Nack(false, true)
				if err != nil {
					rc.logger.Error("failed to send message back to RabbitMQ", err)
				}
			} else {
				err := d.Ack(false)
				if err != nil {
					rc.logger.Error("failed to delete message from RabbitMQ", err)
				}
			}
		}
	}()

	rc.logger.Info("Started consuming", logger.Field{
		Key:   "queue",
		Value: queueName,
	})
	return nil
}
