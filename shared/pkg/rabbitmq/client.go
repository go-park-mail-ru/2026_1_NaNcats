package rabbitmq

//go:generate mockgen -source=client.go -destination=mocks/mock_amqp.go -package=mocks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	rabbitmqErrors "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/errors"
	"github.com/mailru/easyjson"
	"go.opentelemetry.io/otel"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	DlxExchangeName  = "dlx.exchange"
	DlxBindingSuffix = ".failed"
	DlqSuffix        = ".dlq"
)

// Белый список очередей для создания DLX
var queuesWithDLX = map[string]bool{
	"queue.analytics.clickhouse": true,
}

func (rc *RabbitClient) shouldSetupDLX(queueName string) bool {
	return queuesWithDLX[queueName]
}

type AMQPChannel interface {
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error
	QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error
	PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	Qos(prefetchCount, prefetchSize int, global bool) error
	ConsumeWithContext(ctx context.Context, queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	Close() error
}

type RabbitClient struct {
	conn   *amqp.Connection
	ch     AMQPChannel
	logger logger.Logger
}

func NewTestRabbitClient(ch AMQPChannel, logger logger.Logger) *RabbitClient {
	return &RabbitClient{
		ch:     ch,
		logger: logger,
	}
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

// Декларирует DLX, DLQ и связывает их
func (rc *RabbitClient) setupTopology(ctx context.Context, queueName string) error {
	if !rc.shouldSetupDLX(queueName) {
		return nil
	}

	// Декларируем общий Dead Letter Exchange
	err := rc.ch.ExchangeDeclare(
		DlxExchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare DLX: %w", err)
	}

	// Декларируем Dead Letter Queue для конкретной очереди
	dlqName := queueName + DlqSuffix
	_, err = rc.ch.QueueDeclare(
		dlqName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare DLQ [%s]: %w", dlqName, err)
	}

	// Связываем DLQ с DLX по ключу queueName + ".failed"
	routingKey := queueName + DlxBindingSuffix
	err = rc.ch.QueueBind(
		dlqName,
		routingKey,
		DlxExchangeName,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind DLQ to DLX: %w", err)
	}

	return nil
}

func (rc *RabbitClient) PublishJSON(ctx context.Context, queueName string, data any) error {
	if err := rc.setupTopology(ctx, queueName); err != nil {
		return err
	}

	var args amqp.Table
	if rc.shouldSetupDLX(queueName) {
		routingKey := queueName + DlxBindingSuffix
		args = amqp.Table{
			"x-dead-letter-exchange":    DlxExchangeName,
			"x-dead-letter-routing-key": routingKey,
		}
	}

	q, err := rc.ch.QueueDeclare(
		queueName,
		true, false, false, false, args,
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

	headers := make(map[string]interface{})
	InjectContext(ctx, headers)

	err = rc.ch.PublishWithContext(ctx,
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Headers:      headers,
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}

	return nil
}

func (rc *RabbitClient) ConsumeJSON(ctx context.Context, queueName string, handler func(ctx context.Context, body []byte) error) error {
	if err := rc.setupTopology(ctx, queueName); err != nil {
		return err
	}

	var args amqp.Table
	if rc.shouldSetupDLX(queueName) {
		routingKey := queueName + DlxBindingSuffix
		args = amqp.Table{
			"x-dead-letter-exchange":    DlxExchangeName,
			"x-dead-letter-routing-key": routingKey,
		}
	}

	q, err := rc.ch.QueueDeclare(
		queueName,
		true, false, false, false, args,
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
			msgCtx := ExtractContext(context.Background(), d.Headers)

			tracer := otel.Tracer("rabbitmq-consumer")
			msgCtx, span := tracer.Start(msgCtx, "RabbitMQ.Consume "+queueName)

			err := handler(msgCtx, d.Body)
			if err != nil {
				span.RecordError(err)
				rc.logger.Error("failed to process message", err, logger.Field{
					Key:   "queue",
					Value: queueName,
				})

				retryable := rabbitmqErrors.IsRetryable(err)
				if !retryable {
					rc.logger.Warn("non-retryable error encountered, routing to DLQ",
						logger.String("queue", queueName),
						logger.Err(err),
					)
				}

				nackErr := d.Nack(false, retryable)
				if nackErr != nil {
					rc.logger.Error("failed to send Nack to RabbitMQ", nackErr)
				}
				if retryable {
					time.Sleep(1 * time.Second)
				}
			} else {
				err := d.Ack(false)
				if err != nil {
					rc.logger.Error("failed to delete message from RabbitMQ", err)
				}
			}
			span.End()
		}
	}()

	rc.logger.Info("Started consuming", logger.Field{
		Key:   "queue",
		Value: queueName,
	})
	return nil
}
