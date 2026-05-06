package rabbitmq

import (
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/usecase/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewPaymentConsumer(t *testing.T) {
	t.Run("Успешное создание Consumer", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUC := mocks.NewMockPaymentUseCase(ctrl)
		nopLogger := logger.NewNopLogger()

		dummyClient := &rabbitmq.RabbitClient{}

		consumer := NewPaymentConsumer(dummyClient, mockUC, nopLogger)

		assert.NotNil(t, consumer)
		assert.Equal(t, dummyClient, consumer.client)
		assert.Equal(t, mockUC, consumer.usecase)
		assert.Equal(t, nopLogger, consumer.logger)
	})
}
