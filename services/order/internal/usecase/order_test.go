package usecase

import (
	"context"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/domain"
	repoMocks "github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/repository/mocks"
	ucMocks "github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/usecase/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
)

type statusCoder interface {
	GRPCStatus() codes.Code
}

type useCaseDeps struct {
	repo *repoMocks.MockOrderRepository
	addr *ucMocks.MockAddressClient
	cart *ucMocks.MockCartClient
	rest *ucMocks.MockRestaurantClient
	pub  *ucMocks.MockMessagePublisher
}

func setupDeps(ctrl *gomock.Controller) useCaseDeps {
	return useCaseDeps{
		repo: repoMocks.NewMockOrderRepository(ctrl),
		addr: ucMocks.NewMockAddressClient(ctrl),
		cart: ucMocks.NewMockCartClient(ctrl),
		rest: ucMocks.NewMockRestaurantClient(ctrl),
		pub:  ucMocks.NewMockMessagePublisher(ctrl),
	}
}

func TestOrderUseCase_CreateOrder(t *testing.T) {
	type mockInit func(d useCaseDeps)

	ownerID := int64(1)

	tests := []struct {
		name        string
		req         domain.CreateOrderInput
		mockInit    mockInit
		expectedID  string
		expectedErr bool
		errCode     codes.Code
	}{
		{
			name: "Успешное создание заказа (Оплата за всех)",
			req: domain.CreateOrderInput{
				UserID:             1,
				AddressPublicID:    "addr-1",
				RestaurantBranchID: 10,
				RestaurantBrandID:  20,
				DeliveryCost:       100,
				ServiceFee:         50,
				PayForAll:          true,
			},
			mockInit: func(d useCaseDeps) {
				d.cart.EXPECT().GetCart(gomock.Any(), int64(1)).Return(domain.Cart{
					Items: []domain.CartItem{{DishID: 1, Quantity: 2, Price: 500}},
				}, int64(1000), nil)

				d.addr.EXPECT().CheckAddressExists(gomock.Any(), int64(1), "addr-1").Return(nil)
				d.rest.EXPECT().GetRestaurantName(gomock.Any(), int64(10)).Return("KFC", nil)

				d.repo.EXPECT().CreateOrder(gomock.Any(), gomock.Any(), "idem-1").Return("pub-uuid", nil)
				d.pub.EXPECT().PublishJSON(gomock.Any(), events.QueueCartCommands, gomock.Any()).Return(nil)
			},
			expectedID:  "pub-uuid",
			expectedErr: false,
		},
		{
			name: "Ошибка: пустая корзина",
			req:  domain.CreateOrderInput{UserID: 1},
			mockInit: func(d useCaseDeps) {
				d.cart.EXPECT().GetCart(gomock.Any(), int64(1)).Return(domain.Cart{Items: []domain.CartItem{}}, int64(0), nil)
			},
			expectedErr: true,
			errCode:     codes.InvalidArgument,
		},
		{
			name: "Ошибка: нераспределенные предметы при раздельной оплате",
			req: domain.CreateOrderInput{
				UserID:    1,
				PayForAll: false,
			},
			mockInit: func(d useCaseDeps) {
				d.cart.EXPECT().GetCart(gomock.Any(), int64(1)).Return(domain.Cart{
					Items: []domain.CartItem{{DishID: 1, Quantity: 1, Price: 500, OwnerUserID: nil}},
				}, int64(500), nil)

				d.addr.EXPECT().CheckAddressExists(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

				d.rest.EXPECT().GetRestaurantName(gomock.Any(), gomock.Any()).Return("Any Restaurant", nil)
			},
			expectedErr: true,
			errCode:     codes.FailedPrecondition,
		},
		{
			name: "Успешное создание со сплитом",
			req: domain.CreateOrderInput{
				UserID:          1,
				AddressPublicID: "addr-2",
				PayForAll:       false,
				PayerMapping:    map[int64]int64{2: 1},
			},
			mockInit: func(d useCaseDeps) {
				friendID := int64(2)
				d.cart.EXPECT().GetCart(gomock.Any(), int64(1)).Return(domain.Cart{
					Items: []domain.CartItem{
						{DishID: 1, Quantity: 1, Price: 500, OwnerUserID: &ownerID},
						{DishID: 2, Quantity: 1, Price: 300, OwnerUserID: &friendID},
					},
				}, int64(800), nil)

				d.addr.EXPECT().CheckAddressExists(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				d.rest.EXPECT().GetRestaurantName(gomock.Any(), gomock.Any()).Return("Burger King", nil)

				// Проверяем, что создается заказ и публикуется событие
				d.repo.EXPECT().CreateOrder(gomock.Any(), gomock.Any(), "idem-2").Return("pub-uuid-2", nil)
				d.pub.EXPECT().PublishJSON(gomock.Any(), events.QueueCartCommands, gomock.Any()).Return(nil)
			},
			expectedID:  "pub-uuid-2",
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			deps := setupDeps(ctrl)
			tt.mockInit(deps)

			uc := NewOrderUseCase(deps.repo, deps.addr, deps.cart, deps.rest, deps.pub, "http://default-logo", logger.NewNopLogger())

			// Для тестов прокидываем разные ключи идемпотентности в зависимости от кейса
			idemKey := "idem-1"
			if tt.name == "Успешное создание со сплитом" {
				idemKey = "idem-2"
			}

			id, err := uc.CreateOrder(context.Background(), tt.req, idemKey)

			if tt.expectedErr {
				assert.Error(t, err)
				domainErr, ok := err.(statusCoder)
				if ok {
					assert.Equal(t, tt.errCode, domainErr.GRPCStatus())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
}

func TestOrderUseCase_CancelOrder(t *testing.T) {
	type mockInit func(d useCaseDeps)

	tests := []struct {
		name        string
		orderID     string
		userID      int64
		mockInit    mockInit
		expectedErr bool
		errCode     codes.Code
	}{
		{
			name:    "Успешная отмена заказа",
			orderID: "pub-123",
			userID:  1,
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().GetOrderByPublicID(gomock.Any(), "pub-123").Return(domain.Order{
					AdminID: 1, Status: StatusCreated,
				}, nil)
				d.repo.EXPECT().UpdateOrderStatus(gomock.Any(), "pub-123", StatusCancelled).Return(nil)
			},
			expectedErr: false,
		},
		{
			name:    "Ошибка: пользователь не владелец",
			orderID: "pub-123",
			userID:  2,
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().GetOrderByPublicID(gomock.Any(), "pub-123").Return(domain.Order{
					AdminID: 1, Status: StatusCreated,
				}, nil)
			},
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name:    "Ошибка: заказ в терминальном статусе",
			orderID: "pub-123",
			userID:  1,
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().GetOrderByPublicID(gomock.Any(), "pub-123").Return(domain.Order{
					AdminID: 1, Status: StatusFinished,
				}, nil)
			},
			expectedErr: true,
			errCode:     codes.FailedPrecondition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			deps := setupDeps(ctrl)
			tt.mockInit(deps)

			uc := NewOrderUseCase(deps.repo, deps.addr, deps.cart, deps.rest, deps.pub, "", logger.NewNopLogger())

			err := uc.CancelOrder(context.Background(), tt.orderID, tt.userID)

			if tt.expectedErr {
				require.Error(t, err)
				domainErr, ok := err.(statusCoder)
				if ok {
					assert.Equal(t, tt.errCode, domainErr.GRPCStatus())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOrderUseCase_UpdateOrderStatusByPaymentID(t *testing.T) {
	type mockInit func(d useCaseDeps)

	tests := []struct {
		name        string
		paymentID   string
		status      string
		mockInit    mockInit
		expectedErr bool
	}{
		{
			name:        "Пропуск не-succeeded статуса",
			paymentID:   "pay-1",
			status:      "pending",
			mockInit:    func(d useCaseDeps) {}, // Никаких вызовов БД не ожидается
			expectedErr: false,
		},
		{
			name:      "Успех: сплит оплачен, но заказ еще не полностью",
			paymentID: "pay-2",
			status:    "succeeded",
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().UpdateSplitStatusByPaymentID(gomock.Any(), "pay-2", SplitStatusPaid).Return("pub-order", nil)
				d.repo.EXPECT().AreAllSplitsPaid(gomock.Any(), "pub-order").Return(false, nil)
			},
			expectedErr: false,
		},
		{
			name:      "Успех: оплачен последний сплит, весь заказ переходит в Paid",
			paymentID: "pay-3",
			status:    "succeeded",
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().UpdateSplitStatusByPaymentID(gomock.Any(), "pay-3", SplitStatusPaid).Return("pub-order", nil)
				d.repo.EXPECT().AreAllSplitsPaid(gomock.Any(), "pub-order").Return(true, nil)
				d.repo.EXPECT().UpdateOrderStatus(gomock.Any(), "pub-order", StatusPaid).Return(nil)

				// Должен уйти ивент в RabbitMQ
				d.pub.EXPECT().PublishJSON(gomock.Any(), events.QueueGatewayEvents, gomock.Any()).Return(nil)
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			deps := setupDeps(ctrl)
			tt.mockInit(deps)

			uc := NewOrderUseCase(deps.repo, deps.addr, deps.cart, deps.rest, deps.pub, "", logger.NewNopLogger())

			err := uc.UpdateOrderStatusByPaymentID(context.Background(), tt.paymentID, tt.status, "idem-key")

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOrderUseCase_ProcessSagaReply(t *testing.T) {
	type mockInit func(d useCaseDeps)

	tests := []struct {
		name        string
		reply       events.SagaReply
		mockInit    mockInit
		expectedErr bool
	}{
		{
			name: "Ошибка саги на шаге PAYMENT -> заказ в payment_ready для ретрая",
			reply: events.SagaReply{
				OrderID: "pub-123", Step: "PAYMENT", Status: events.StatusError, SplitID: "split-1",
			},
			mockInit: func(d useCaseDeps) {
				// Платёж не создался, но заказ валиден: переводим в payment_ready,
				// чтобы фронт мог предложить повторную оплату/другую карту.
				d.repo.EXPECT().UpdateOrderStatus(gomock.Any(), "pub-123", StatusPaymentReady).Return(nil)
				d.repo.EXPECT().UpdateSplitStatus(gomock.Any(), "split-1", SplitStatusFailed).Return(nil)
				// Событие для фронта через gateway-очередь.
				d.pub.EXPECT().PublishJSON(gomock.Any(), events.QueueGatewayEvents, gomock.Any()).Return(nil)
			},
			expectedErr: false,
		},
		{
			name: "Успешный шаг CART -> отправка команд на оплату",
			reply: events.SagaReply{
				OrderID: "pub-123", Step: "CART", Status: events.StatusSuccess,
			},
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().GetOrderByPublicID(gomock.Any(), "pub-123").Return(domain.Order{
					PublicID: "pub-123",
					Splits:   []domain.OrderSplit{{ID: "split-1", UserID: 1, Amount: 1000}},
				}, nil)
				d.repo.EXPECT().UpdateOrderStatus(gomock.Any(), "pub-123", StatusCartLocked).Return(nil)
				d.pub.EXPECT().PublishJSON(gomock.Any(), events.QueuePaymentCommands, gomock.Any()).Return(nil)
			},
			expectedErr: false,
		},
		{
			name: "Успешный шаг PAYMENT -> сохранение ID и отправка в Gateway",
			reply: events.SagaReply{
				OrderID: "pub-123", Step: "PAYMENT", Status: events.StatusSuccess, SplitID: "split-1", PaymentID: "yoo-123", PaymentURL: "http://pay",
			},
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().SetSplitYookassaID(gomock.Any(), "split-1", "yoo-123").Return(nil)
				d.pub.EXPECT().PublishJSON(gomock.Any(), events.QueueGatewayEvents, gomock.Any()).Return(nil)
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			deps := setupDeps(ctrl)
			tt.mockInit(deps)

			uc := NewOrderUseCase(deps.repo, deps.addr, deps.cart, deps.rest, deps.pub, "", logger.NewNopLogger())

			err := uc.ProcessSagaReply(context.Background(), tt.reply)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOrderUseCase_GetOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	deps := setupDeps(ctrl)
	uc := NewOrderUseCase(deps.repo, deps.addr, deps.cart, deps.rest, deps.pub, "http://default-logo", logger.NewNopLogger())

	t.Run("Успешное получение с обогащением логотипами", func(t *testing.T) {
		deps.repo.EXPECT().GetOrdersByUserID(gomock.Any(), int64(1)).Return([]domain.Order{
			{ID: 1, RestaurantBrandID: 10},
			{ID: 2, RestaurantBrandID: 20},
		}, nil)

		deps.rest.EXPECT().GetLogosByBrandIDs(gomock.Any(), gomock.Any()).Return(map[int64]string{
			10: "http://logo-10",
		}, nil)

		orders, err := uc.GetOrders(context.Background(), 1)

		assert.NoError(t, err)
		assert.Len(t, orders, 2)
		assert.Equal(t, "http://logo-10", orders[0].RestaurantLogoURL)
		assert.Equal(t, "http://default-logo", orders[1].RestaurantLogoURL) // Фолбэк на дефолт
	})

	t.Run("Пустой список заказов", func(t *testing.T) {
		deps.repo.EXPECT().GetOrdersByUserID(gomock.Any(), int64(1)).Return([]domain.Order{}, nil)

		orders, err := uc.GetOrders(context.Background(), 1)

		assert.NoError(t, err)
		assert.Len(t, orders, 0)
	})
}
