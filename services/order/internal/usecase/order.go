package usecase

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
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

	SplitStatusPending   = "pending"
	SplitStatusPaid      = "paid"
	SplitStatusFailed    = "failed"
	SplitStatusCancelled = "cancelled"
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
	CreateOrder(ctx context.Context, req domain.CreateOrderInput, idempotencyKey string) (string, error)
	GetOrders(ctx context.Context, userID int64) ([]domain.Order, error)
	UpdateOrderStatusByPaymentID(ctx context.Context, paymentID string, status string, idempotencyKey string) error
	ProcessSagaReply(ctx context.Context, reply events.SagaReply) error
	PayForFriend(ctx context.Context, splitID string, adminID int64, paymentMethodID, idempotencyKey string) error
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

func (o *orderUseCase) CreateOrder(ctx context.Context, req domain.CreateOrderInput, idempotencyKey string) (string, error) {
	cart, cartTotalCost, err := o.cartClient.GetCart(ctx, req.UserID)
	if err != nil {
		return "", errutil.Wrap("INTERNAL_ERROR", "failed to get cart", err, codes.Internal)
	}
	if len(cart.Items) == 0 {
		return "", errutil.New("CART_EMPTY", "cart is empty", codes.InvalidArgument)
	}

	err = o.addressClient.CheckAddressExists(ctx, req.UserID, req.AddressPublicID)
	if err != nil {
		return "", errutil.Wrap("INVALID_ADDRESS", "address invalid", err, codes.InvalidArgument)
	}

	resName, err := o.restaurantClient.GetRestaurantName(ctx, req.RestaurantBranchID)
	if err != nil {
		return "", errutil.Wrap("INTERNAL_ERROR", "failed to get restaurant", err, codes.Internal)
	}

	finalTotalCost := cartTotalCost + req.DeliveryCost + req.ServiceFee

	userDebts := make(map[int64]int64)

	if req.PayForAll {
		userDebts[req.UserID] = finalTotalCost
	} else {
		for _, item := range cart.Items {
			if item.OwnerUserID == nil {
				return "", errutil.New("UNASSIGNED_ITEMS", "cannot checkout: cart has unassigned items", codes.FailedPrecondition)
			}
			userDebts[*item.OwnerUserID] += item.Price * int64(item.Quantity)
		}

		for targetID, payerID := range req.PayerMapping {
			if debt, exists := userDebts[targetID]; exists {
				userDebts[payerID] += debt
				delete(userDebts, targetID)
			}
		}

		userDebts[req.UserID] += req.DeliveryCost + req.ServiceFee
	}

	items := make([]domain.OrderDish, 0, len(cart.Items))
	for _, item := range cart.Items {
		items = append(items, domain.OrderDish{
			DishID:      item.DishID,
			Quantity:    item.Quantity,
			Price:       item.Price,
			OwnerUserID: item.OwnerUserID,
		})
	}

	splits := make([]domain.OrderSplit, 0, len(userDebts))
	for uid, amount := range userDebts {
		if amount > 0 {
			splits = append(splits, domain.OrderSplit{
				ID:     uuid.New().String(),
				UserID: uid,
				Amount: amount,
				Status: SplitStatusPending,
			})
		}
	}

	order := domain.Order{
		AdminID:            req.UserID,
		RestaurantBranchID: req.RestaurantBranchID,
		RestaurantBrandID:  req.RestaurantBrandID,
		RestaurantName:     resName,
		ClientAddressID:    req.AddressPublicID,
		TotalCost:          finalTotalCost,
		Status:             StatusCreated,
		Items:              items,
		Splits:             splits,
	}

	orderPublicID, err := o.orderRepo.CreateOrder(ctx, order, idempotencyKey)
	if err != nil {
		return "", errutil.Wrap("INTERNAL_ERROR", "failed to save order", err, codes.Internal)
	}

	// Запускаем Сагу
	cmd := events.SagaCommand{
		OrderID:        orderPublicID,
		UserID:         req.UserID,
		Action:         events.CommandLockCart,
		IdempotencyKey: idempotencyKey,
	}
	_ = o.rabbitPublisher.PublishJSON(ctx, events.QueueCartCommands, cmd)

	return orderPublicID, nil
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
	if status != "succeeded" {
		return nil
	}

	orderPublicID, err := o.orderRepo.UpdateSplitStatusByPaymentID(ctx, paymentID, SplitStatusPaid)
	if err != nil {
		return err
	}

	allPaid, err := o.orderRepo.AreAllSplitsPaid(ctx, orderPublicID)
	if err != nil {
		return err
	}

	if allPaid {
		err = o.orderRepo.UpdateOrderStatus(ctx, orderPublicID, StatusInProgress)
		if err != nil {
			return err
		}

		// TODO: можно отправить событие в RestaurantService для начала готовки
		o.logger.Info("All splits paid! Order is in progress", logger.String("order_id", orderPublicID))
	}

	return nil
}

func (o *orderUseCase) PayForFriend(ctx context.Context, splitID string, adminID int64, paymentMethodID, idempotencyKey string) error {
	// Меняем плательщика в БД (только если сплит не оплачен)
	err := o.orderRepo.UpdateSplitPayer(ctx, splitID, adminID)
	if err != nil {
		return errutil.Wrap("CANNOT_REASSIGN", "failed to reassign split", err, codes.InvalidArgument)
	}

	split, err := o.orderRepo.GetSplitByID(ctx, splitID)
	if err != nil {
		return err
	}

	// Генерируем команду для платежки
	payCmd := events.SagaCommand{
		OrderID:         fmt.Sprintf("%d", split.OrderID), // Временный костыль, лучше вытащить PublicID
		SplitID:         split.ID,
		UserID:          adminID,
		Action:          events.CommandCreatePayment,
		Amount:          split.Amount,
		PaymentMethodID: paymentMethodID,
		IdempotencyKey:  idempotencyKey,
	}

	return o.rabbitPublisher.PublishJSON(ctx, events.QueuePaymentCommands, payCmd)
}

func (o *orderUseCase) ProcessSagaReply(ctx context.Context, reply events.SagaReply) error {
	if reply.Status == events.StatusError {
		_ = o.orderRepo.UpdateOrderStatus(ctx, reply.OrderID, StatusFailed)
		if reply.SplitID != "" {
			_ = o.orderRepo.UpdateSplitStatus(ctx, reply.SplitID, SplitStatusFailed)
		}
		// Откат корзины
		if reply.Step == "PAYMENT" {
			_ = o.rabbitPublisher.PublishJSON(ctx, events.QueueCartCommands, events.SagaCommand{
				OrderID: reply.OrderID, Action: events.CommandUnlockCart, IdempotencyKey: reply.OrderID + "_compensate",
			})
		}
		return nil
	}

	switch reply.Step {
	case "CART":
		order, err := o.orderRepo.GetOrderByPublicID(ctx, reply.OrderID)
		if err != nil {
			return err
		}

		_ = o.orderRepo.UpdateOrderStatus(ctx, reply.OrderID, StatusCartLocked)

		for _, split := range order.Splits {
			payCmd := events.SagaCommand{
				OrderID:        order.PublicID,
				SplitID:        split.ID,
				UserID:         split.UserID,
				Action:         events.CommandCreatePayment,
				Amount:         split.Amount,
				IdempotencyKey: split.ID + "_payment",
			}
			_ = o.rabbitPublisher.PublishJSON(ctx, events.QueuePaymentCommands, payCmd)
		}

	case "PAYMENT":
		_ = o.orderRepo.SetSplitYookassaID(ctx, reply.SplitID, reply.PaymentID)

		gatewayEvent := events.GatewayEvent{
			OrderID:    reply.OrderID,
			SplitID:    reply.SplitID,
			UserID:     reply.UserID,
			Status:     StatusPaymentReady,
			PaymentURL: reply.PaymentURL,
		}
		_ = o.rabbitPublisher.PublishJSON(ctx, events.QueueGatewayEvents, gatewayEvent)
	}

	return nil
}
