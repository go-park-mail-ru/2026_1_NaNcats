package rabbitmq

import (
	"context"
	"errors"
	"testing"

	ucMocks "github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/usecase/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq"
	rabbitmqErrors "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/errors"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"
	"github.com/mailru/easyjson"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewAnalyticsConsumer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uc := ucMocks.NewMockAnalyticsUseCase(ctrl)
	nop := logger.NewNopLogger()
	rabbit := &rabbitmq.RabbitClient{}

	c := NewAnalyticsConsumer(rabbit, uc, nop)

	assert.NotNil(t, c)
	assert.Equal(t, rabbit, c.rabbit)
	assert.Equal(t, uc, c.uc)
	assert.Equal(t, nop, c.logger)
}

func TestAnalyticsConsumer_handleMessage(t *testing.T) {
	validEvent := events.AnalyticsOrderEvent{
		OrderPublicID: "ord-xyz",
		Status:        "paid",
		EventTime:     1700000000000,
		RestaurantID:  100,
	}
	validBody, err := easyjson.Marshal(validEvent)
	assert.NoError(t, err)

	tests := []struct {
		name              string
		body              []byte
		mockInit          func(uc *ucMocks.MockAnalyticsUseCase)
		expectErr         bool
		expectPermanent   bool
		expectErrContains string
	}{
		{
			name: "Корректное сообщение прокидывается в usecase",
			body: validBody,
			mockInit: func(uc *ucMocks.MockAnalyticsUseCase) {
				uc.EXPECT().
					ProcessEvent(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, e events.AnalyticsOrderEvent) error {
						assert.Equal(t, "ord-xyz", e.OrderPublicID)
						assert.Equal(t, "paid", e.Status)
						return nil
					})
			},
		},
		{
			name:     "Битый JSON отбрасывается как PermanentError (не вернётся в очередь)",
			body:     []byte("{not json"),
			mockInit: func(uc *ucMocks.MockAnalyticsUseCase) {},
			// usecase не должен быть вызван
			expectErr:         true,
			expectPermanent:   true,
			expectErrContains: "permanent error",
		},
		{
			name: "Ошибка usecase пробрасывается как обычная (для retry)",
			body: validBody,
			mockInit: func(uc *ucMocks.MockAnalyticsUseCase) {
				uc.EXPECT().ProcessEvent(gomock.Any(), gomock.Any()).
					Return(errors.New("clickhouse down"))
			},
			expectErr:         true,
			expectPermanent:   false,
			expectErrContains: "clickhouse down",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			uc := ucMocks.NewMockAnalyticsUseCase(ctrl)
			tt.mockInit(uc)

			c := NewAnalyticsConsumer(&rabbitmq.RabbitClient{}, uc, logger.NewNopLogger())
			err := c.handleMessage(context.Background(), tt.body)

			if !tt.expectErr {
				assert.NoError(t, err)
				return
			}

			assert.Error(t, err)
			if tt.expectErrContains != "" {
				assert.Contains(t, err.Error(), tt.expectErrContains)
			}
			var perm *rabbitmqErrors.PermanentError
			isPermanent := errors.As(err, &perm)
			assert.Equal(t, tt.expectPermanent, isPermanent, "PermanentError expectation mismatch")
		})
	}
}
