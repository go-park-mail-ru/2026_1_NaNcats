package rabbitmq

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/mocks"

	amqp "github.com/rabbitmq/amqp091-go"
)

// dummyLogger заглушка для логгера, чтобы не усложнять мок (если logger это интерфейс)
type dummyLogger struct{ logger.Logger }

func (d *dummyLogger) Warn(msg string, fields ...logger.Field)             {}
func (d *dummyLogger) Error(msg string, err error, fields ...logger.Field) {}
func (d *dummyLogger) Info(msg string, fields ...logger.Field)             {}

func TestRabbitClient_PublishJSON(t *testing.T) {
	type mockInit func(m *mocks.MockAMQPChannel)

	tests := []struct {
		name        string
		queueName   string
		data        any
		mockInit    mockInit
		expectError bool
	}{
		{
			name:      "Успешная публикация сообщения",
			queueName: "test.queue",
			data:      map[string]string{"key": "value"}, // Будет использован fallback на json.Marshal
			mockInit: func(m *mocks.MockAMQPChannel) {
				// Ожидаем объявление очереди
				m.EXPECT().QueueDeclare("test.queue", true, false, false, false, nil).
					Return(amqp.Queue{Name: "test.queue"}, nil)

				// Ожидаем успешную публикацию
				m.EXPECT().PublishWithContext(gomock.Any(), "", "test.queue", false, false, gomock.Any()).
					Return(nil)
			},
			expectError: false,
		},
		{
			name:      "Ошибка при объявлении очереди",
			queueName: "error.queue",
			data:      map[string]string{"key": "value"},
			mockInit: func(m *mocks.MockAMQPChannel) {
				m.EXPECT().QueueDeclare("error.queue", true, false, false, false, nil).
					Return(amqp.Queue{}, errors.New("declare failed"))
			},
			expectError: true,
		},
		{
			name:      "Ошибка при публикации в канал",
			queueName: "test.queue",
			data:      map[string]string{"key": "value"},
			mockInit: func(m *mocks.MockAMQPChannel) {
				m.EXPECT().QueueDeclare("test.queue", true, false, false, false, nil).
					Return(amqp.Queue{Name: "test.queue"}, nil)

				m.EXPECT().PublishWithContext(gomock.Any(), "", "test.queue", false, false, gomock.Any()).
					Return(errors.New("publish failed"))
			},
			expectError: true,
		},
		{
			name:      "Ошибка маршалинга (невалидные данные)",
			queueName: "test.queue",
			data:      make(chan int), // chan нельзя сериализовать в JSON
			mockInit: func(m *mocks.MockAMQPChannel) {
				m.EXPECT().QueueDeclare("test.queue", true, false, false, false, nil).
					Return(amqp.Queue{Name: "test.queue"}, nil)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockChan := mocks.NewMockAMQPChannel(ctrl)
			tt.mockInit(mockChan)

			client := NewTestRabbitClient(mockChan, &dummyLogger{})
			err := client.PublishJSON(context.Background(), tt.queueName, tt.data)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRabbitClient_ConsumeJSON(t *testing.T) {
	type mockInit func(m *mocks.MockAMQPChannel)

	tests := []struct {
		name        string
		queueName   string
		mockInit    mockInit
		expectError bool
	}{
		{
			name:      "Успешная инициализация консьюмера",
			queueName: "consume.queue",
			mockInit: func(m *mocks.MockAMQPChannel) {
				m.EXPECT().QueueDeclare("consume.queue", true, false, false, false, nil).
					Return(amqp.Queue{Name: "consume.queue"}, nil)

				m.EXPECT().Qos(10, 0, false).Return(nil)

				// Возвращаем пустой канал, чтобы горутина не висела и тест быстро завершился
				emptyDelivery := make(chan amqp.Delivery)
				close(emptyDelivery)
				m.EXPECT().ConsumeWithContext(gomock.Any(), "consume.queue", "", false, false, false, false, nil).
					Return(emptyDelivery, nil)
			},
			expectError: false,
		},
		{
			name:      "Ошибка при установке QoS",
			queueName: "consume.queue",
			mockInit: func(m *mocks.MockAMQPChannel) {
				m.EXPECT().QueueDeclare("consume.queue", true, false, false, false, nil).
					Return(amqp.Queue{Name: "consume.queue"}, nil)

				m.EXPECT().Qos(10, 0, false).Return(errors.New("qos failed"))
			},
			expectError: true,
		},
		{
			name:      "Ошибка регистрации консьюмера",
			queueName: "consume.queue",
			mockInit: func(m *mocks.MockAMQPChannel) {
				m.EXPECT().QueueDeclare("consume.queue", true, false, false, false, nil).
					Return(amqp.Queue{Name: "consume.queue"}, nil)

				m.EXPECT().Qos(10, 0, false).Return(nil)

				m.EXPECT().ConsumeWithContext(gomock.Any(), "consume.queue", "", false, false, false, false, nil).
					Return(nil, errors.New("consume failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockChan := mocks.NewMockAMQPChannel(ctrl)
			tt.mockInit(mockChan)

			client := NewTestRabbitClient(mockChan, &dummyLogger{})
			err := client.ConsumeJSON(context.Background(), tt.queueName, func(ctx context.Context, body []byte) error {
				return nil
			})

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
