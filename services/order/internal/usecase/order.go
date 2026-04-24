package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	StatusCreated      = "created"
	StatusCartLocked   = "cart_locked"
	StatusPaymentReady = "payment_ready"
	StatusPaid         = "paid"
	StatusInProgress   = "in_progress"
	StatusWaiting      = "waiting"
	StatusDelivering   = "delivering"
	StatusFinished     = "finished"
	StatusCancelled    = "cancelled"
	StatusFailed       = "failed"
)

//go:generate mockgen -destination=mocks/cart_client_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/usecase CartClient
type CartClient interface {
	GetCart(ctx context.Context, userID int64) (domain.Cart, int64, error)
}

//go:generate mockgen -destination=mocks/address_client_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/usecase AddressClient
type AddressClient interface {
	CheckAddressExists(ctx context.Context, userID int64, addressPublicID string) error
}

//go:generate mockgen -destination=mocks/restaurant_client_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/usecase RestaurantClient
type RestaurantClient interface {
	GetRestaurantName(ctx context.Context, branchID int64) (string, error)
	GetLogosByBrandIDs(ctx context.Context, brandIDs []int64) (map[int64]string, error)
}

type MessagePublisher interface {
	PublishJSON(ctx context.Context, queueName string, data any) error
}

//go:generate mockgen -destination=mocks/order_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/order OrderUseCase
type OrderUseCase interface {
	CreateOrder(ctx context.Context, userID int64, req domain.CreateOrderInput, idempotencyKey string) (string, string, error)
	GetOrders(ctx context.Context, userID int64) ([]domain.Order, error)
	UpdateOrderStatusByPaymentID(ctx context.Context, paymentID string, status string, idempotencyKey string) error
	ProcessSagaReply(ctx context.Context, reply events.SagaReply) error
}

type orderUseCase struct {
	orderRepo                repository.OrderRepository
	addressClient            AddressClient
	cartClient               CartClient
	restaurantClient         RestaurantClient
	rabbitPublisher          MessagePublisher
	defaultRestaurantLogoURL string
	logger                   logger.Logger
}

func NewOrderUseCase(
	or repository.OrderRepository,
	ac AddressClient,
	cc CartClient,
	rc RestaurantClient,
	rp MessagePublisher,
	drlurl string,
	l logger.Logger,
) OrderUseCase {
	return &orderUseCase{
		orderRepo:                or,
		addressClient:            ac,
		cartClient:               cc,
		restaurantClient:         rc,
		rabbitPublisher:          rp,
		defaultRestaurantLogoURL: drlurl,
		logger:                   l,
	}
}

func (o *orderUseCase) CreateOrder(ctx context.Context, userID int64, req domain.CreateOrderInput, idempotencyKey string) (string, string, error) {
	// Читаем корзину
	cart, cartTotalCost, err := o.cartClient.GetCart(ctx, userID)
	if err != nil {
		return "", "", errutil.Wrap("INTERNAL_SERVER_ERROR", "failed to get cart", err, codes.Internal)
	}
	if len(cart.Items) == 0 {
		return "", "", errutil.New("CART_EMPTY", "cart is empty", codes.InvalidArgument)
	}

	// Проверяем адрес
	err = o.addressClient.CheckAddressExists(ctx, userID, req.AddressPublicID)
	if err != nil {
		st, ok := status.FromError(err)

		if ok && st.Code() == codes.NotFound {
			return "", "", errutil.New("ADDRESS_NOT_FOUND_OR_INVALID", "user provided bad address", codes.NotFound)
		}

		return "", "", fmt.Errorf("address service internal error: %w", err)
	}

	// Собираем блюда
	items := make([]domain.OrderDish, 0, len(cart.Items))
	for _, item := range cart.Items {
		items = append(items, domain.OrderDish{
			DishID:   item.DishID,
			Quantity: item.Quantity,
			Price:    item.Price,
		})
	}

	// Считаем итоговую сумму
	/* // TODO:
	/  // 1) Реализовать расчет стоимости на основе расстояния между рестиком и адресом пользователя
	/  // 2) Попробовать интегрировать динамичную цену при помощи API погоды
	*/
	finalTotalCost := cartTotalCost + req.DeliveryCost + req.ServiceFee

	// Собираем инфу о ресторане
	resName, err := o.restaurantClient.GetRestaurantName(ctx, req.RestaurantBranchID)
	if err != nil {
		return "", "", errutil.Wrap("INTERNAL_SERVER_ERROR", "failed to get restaurant info", err, codes.Internal)
	}

	// Записываем в БД со статусом "created" (см. файл миграции)
	order := domain.Order{
		ClientID:           userID,
		RestaurantBranchID: req.RestaurantBranchID,
		RestaurantBrandID:  req.RestaurantBrandID,
		RestaurantName:     resName,
		ClientAddressID:    req.AddressPublicID,
		TotalCost:          finalTotalCost,
		Status:             StatusCreated,
		Items:              items,
	}

	orderPublicID, err := o.orderRepo.CreateOrder(ctx, order, idempotencyKey)
	if err != nil {
		return "", "", errutil.Wrap("INTERNAL_SERVER_ERROR", "failed to create order in db", err, codes.Internal)
	}

	// Запуск саги

	// Формируем команду для корзины
	cmd := events.SagaCommand{
		OrderID:        orderPublicID,
		UserID:         userID,
		Action:         events.CommandLockCart,
		IdempotencyKey: idempotencyKey,
		Amount:         finalTotalCost,
	}

	// Публикуем в RabbitMQ
	err = o.rabbitPublisher.PublishJSON(ctx, events.QueueCartCommands, cmd)
	if err != nil {
		// Компенсирующая операция
		cancelErr := o.orderRepo.UpdateOrderStatus(ctx, orderPublicID, StatusFailed)
		if cancelErr != nil {
			o.logger.Error("failed while compensating operation", cancelErr)
		}
		return "", "", errutil.Wrap("INTERNAL_SERVER_ERROR", "broker unavailable, order cancelled", err, codes.Internal)
	}

	return orderPublicID, "", nil
}

func (o *orderUseCase) GetOrders(ctx context.Context, userID int64) ([]domain.Order, error) {
	orders, err := o.orderRepo.GetOrdersByUserID(ctx, userID)
	if err != nil {
		return []domain.Order{}, errutil.Wrap("INTERNAL_SERVER_ERROR", "failed to get orders", err, codes.Internal)
	}

	if len(orders) == 0 {
		return orders, nil
	}

	// Собираем уникальные ID ресторанов
	brandIDs := make([]int64, 0)
	seen := make(map[int64]bool)
	for _, ord := range orders {
		if !seen[ord.RestaurantBrandID] {
			brandIDs = append(brandIDs, ord.RestaurantBrandID)
			seen[ord.RestaurantBrandID] = true
		}
	}

	logos, err := o.restaurantClient.GetLogosByBrandIDs(ctx, brandIDs)
	if err != nil {
		o.logger.Error("failed to get logos of restaurants", err)
	}

	for i := range orders {
		logo, ok := logos[orders[i].RestaurantBrandID]
		if ok && logo != "" {
			orders[i].RestaurantLogoURL = logo
		} else {
			orders[i].RestaurantLogoURL = o.defaultRestaurantLogoURL
		}
	}

	return orders, nil
}

func (o *orderUseCase) UpdateOrderStatusByPaymentID(ctx context.Context, paymentID string, status string, idempotencyKey string) error {
	var newStatus string
	switch status {
	case "succeeded", "finished":
		newStatus = "in_progress"
	case "canceled":
		newStatus = "canceled"
	default:
		return errutil.New("UNKNOWN_STATUS", "unknown payment status", codes.InvalidArgument)
	}

	return o.orderRepo.UpdateStatusByPaymentID(ctx, paymentID, newStatus)
}

func (o *orderUseCase) ProcessSagaReply(ctx context.Context, reply events.SagaReply) error {
	if reply.Status == events.StatusError {
		o.logger.Error("Saga step failed", errors.New(reply.ErrorMessage), logger.String("step", reply.Step))

		err := o.orderRepo.UpdateOrderStatus(ctx, reply.OrderID, StatusFailed)
		if err != nil {
			o.logger.Error("failed to update order status while handling saga step error", err, logger.String("order_id", reply.OrderID))
		}

		if reply.Step == "PAYMENT" {
			compCmd := events.SagaCommand{
				OrderID:        reply.OrderID,
				Action:         events.CommandUnlockCart,
				IdempotencyKey: reply.OrderID + "_compensate_cart",
			}
			// Компенсирующая операция
			cancelErr := o.rabbitPublisher.PublishJSON(ctx, events.QueueCartCommands, compCmd)
			if cancelErr != nil {
				o.logger.Error("failed while compensating operation", cancelErr)
			}
		}

		gatewayEvent := events.GatewayEvent{
			OrderID: reply.OrderID,
			Status:  StatusFailed,
			Error:   reply.ErrorMessage,
		}
		err = o.rabbitPublisher.PublishJSON(ctx, events.QueueGatewayEvents, gatewayEvent)
		if err != nil {
			o.logger.Error("failed while compensating operation", err)
		}

		return nil
	}

	// Успешные шаги
	switch reply.Step {
	case "CART":
		// Сейчас корзина залочена

		order, err := o.orderRepo.GetOrderByPublicID(ctx, reply.OrderID, reply.UserID)
		if err != nil {
			return fmt.Errorf("failed to get order for payment: %w", err)
		}

		// Обновляем статус
		err = o.orderRepo.UpdateOrderStatus(ctx, reply.OrderID, StatusCartLocked)
		if err != nil {
			return err
		}

		payCmd := events.SagaCommand{
			OrderID:         reply.OrderID,
			UserID:          order.ClientID,
			Action:          events.CommandCreatePayment,
			Amount:          order.TotalCost,
			PaymentMethodID: order.PaymentMethodID,
			IdempotencyKey:  reply.OrderID + "_payment",
		}
		return o.rabbitPublisher.PublishJSON(ctx, events.QueuePaymentCommands, payCmd)

	case "PAYMENT":
		// Платеж создан

		// Сохраняем ID платежа и меняем статус
		err := o.orderRepo.SetYookassaID(ctx, reply.OrderID, reply.PaymentID)
		if err != nil {
			return err
		}
		err = o.orderRepo.UpdateOrderStatus(ctx, reply.OrderID, StatusPaymentReady)
		if err != nil {
			return err
		}

		gatewayEvent := events.GatewayEvent{
			OrderID:    reply.OrderID,
			Status:     StatusPaymentReady,
			PaymentURL: reply.PaymentURL,
		}
		return o.rabbitPublisher.PublishJSON(ctx, events.QueueGatewayEvents, gatewayEvent)

	default:
		o.logger.Warn("Unknown saga step", logger.String("step", reply.Step))
	}

	return nil
}
