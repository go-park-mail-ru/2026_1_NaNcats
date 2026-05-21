package autoadvance

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/infrastructure/autoadvance/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/repository"
	repoMocks "github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/repository/mocks"
	ucMocks "github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/usecase/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -destination=mocks/autoadvance_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/infrastructure/autoadvance Publisher

func TestRunner_tick(t *testing.T) {
	type mockInit func(repoMock *repoMocks.MockOrderRepository, pubMock *mocks.MockPublisher)

	tests := []struct {
		name     string
		mockInit mockInit
	}{
		{
			name: "Успешное продвижение нескольких заказов",
			mockInit: func(r *repoMocks.MockOrderRepository, p *mocks.MockPublisher) {
				orders := []domain.Order{
					{PublicID: "order-1", Status: "paid"},
					{PublicID: "order-2", Status: "in_progress"},
					{PublicID: "order-3", Status: "finished"},
				}

				r.EXPECT().GetOrdersByStatuses(gomock.Any(), sourceStatuses()).Return(orders, nil)

				r.EXPECT().UpdateOrderStatus(gomock.Any(), "order-1", "in_progress", "paid").Return(nil)
				p.EXPECT().PublishJSON(gomock.Any(), events.QueueGatewayEvents, events.GatewayEvent{
					OrderID: "order-1",
					Status:  "in_progress",
				}).Return(nil)

				r.EXPECT().UpdateOrderStatus(gomock.Any(), "order-2", "waiting", "in_progress").Return(nil)
				p.EXPECT().PublishJSON(gomock.Any(), events.QueueGatewayEvents, events.GatewayEvent{
					OrderID: "order-2",
					Status:  "waiting",
				}).Return(nil)
			},
		},
		{
			name: "Ошибка при получении заказов (прерывание tick)",
			mockInit: func(r *repoMocks.MockOrderRepository, p *mocks.MockPublisher) {
				r.EXPECT().GetOrdersByStatuses(gomock.Any(), sourceStatuses()).Return(nil, errors.New("db timeout"))
			},
		},
		{
			name: "Ошибка при обновлении статуса (пропуск публикации)",
			mockInit: func(r *repoMocks.MockOrderRepository, p *mocks.MockPublisher) {
				orders := []domain.Order{
					{PublicID: "order-error", Status: "waiting"},
				}

				r.EXPECT().GetOrdersByStatuses(gomock.Any(), sourceStatuses()).Return(orders, nil)
				r.EXPECT().UpdateOrderStatus(gomock.Any(), "order-error", "delivering", "waiting").Return(errors.New("deadlock"))
			},
		},
		{
			name: "Конкурентное изменение статуса (ErrStateChanged)",
			mockInit: func(r *repoMocks.MockOrderRepository, p *mocks.MockPublisher) {
				orders := []domain.Order{
					{PublicID: "order-concurrent", Status: "paid"},
				}

				r.EXPECT().GetOrdersByStatuses(gomock.Any(), sourceStatuses()).Return(orders, nil)
				r.EXPECT().UpdateOrderStatus(gomock.Any(), "order-concurrent", "in_progress", "paid").Return(repository.ErrStateChanged)
			},
		},
		{
			name: "Нет заказов для продвижения",
			mockInit: func(r *repoMocks.MockOrderRepository, p *mocks.MockPublisher) {
				r.EXPECT().GetOrdersByStatuses(gomock.Any(), sourceStatuses()).Return([]domain.Order{}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repoMock := repoMocks.NewMockOrderRepository(ctrl)
			pubMock := mocks.NewMockPublisher(ctrl)
			tt.mockInit(repoMock, pubMock)

			ucMock := ucMocks.NewMockOrderUseCase(ctrl)

			runner := New(repoMock, pubMock, 1*time.Millisecond, logger.NewNopLogger(), ucMock)

			assert.NotPanics(t, func() {
				runner.tick(context.Background())
			})
		})
	}
}
