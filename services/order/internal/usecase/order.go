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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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

//go:generate mockgen -destination=mocks/order_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/usecase OrderUseCase,CartClient,AddressClient,RestaurantClient,MessagePublisher
//go:generate gowrap gen -i OrderUseCase -t ../../../../shared/templates/tracing.tmpl -o order_tracing_mw.go -v TracerName=order-service

type CartClient interface {
	GetCart(ctx context.Context, userID int64) (domain.Cart, int64, error)
}

type AddressClient interface {
	CheckAddressExists(ctx context.Context, userID int64, addressPublicID string) error
}

type RestaurantClient interface {
	GetRestaurantName(ctx context.Context, branchID int64) (string, error)
	GetLogosByBrandIDs(ctx context.Context, brandIDs []int64) (map[int64]string, error)
}

type MessagePublisher interface {
	PublishJSON(ctx context.Context, queueName string, data any) error
}

type OrderUseCase interface {
	CreateOrder(ctx context.Context, req domain.CreateOrderInput, idempotencyKey string) (string, error)
	GetOrders(ctx context.Context, userID int64) ([]domain.Order, error)
	UpdateOrderStatusByPaymentID(ctx context.Context, paymentID string, status string, idempotencyKey string) error
	ProcessSagaReply(ctx context.Context, reply events.SagaReply) error
	PayForFriend(ctx context.Context, splitID string, adminID int64, paymentMethodID, idempotencyKey string) error
	GetOrderPaymentID(ctx context.Context, orderPublicID string, userID int64) (string, error)
	CancelOrder(ctx context.Context, orderPublicID string, userID int64) error
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
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("user.id", req.UserID),
		attribute.Int64("restaurant.brand_id", req.RestaurantBrandID),
		attribute.String("address.public_id", req.AddressPublicID),
		attribute.Bool("order.pay_for_all", req.PayForAll),
	)

	cart, cartTotalCost, err := o.cartClient.GetCart(ctx, req.UserID)
	if err != nil {
		return "", errutil.Internal("failed to fetch cart for order", err)
	}

	if len(cart.Items) == 0 {
		span.AddEvent("cart_empty_abort")
		return "", errutil.New("CART_EMPTY", "cart is empty", codes.InvalidArgument)
	}

	err = o.addressClient.CheckAddressExists(ctx, req.UserID, req.AddressPublicID)
	if err != nil {
		return "", errutil.Wrap("INVALID_ADDRESS", "address invalid", err, codes.InvalidArgument)
	}

	resName, err := o.restaurantClient.GetRestaurantName(ctx, req.RestaurantBranchID)
	if err != nil {
		return "", errutil.Internal("failed to fetch restaurant name", err)
	}

	finalTotalCost := cartTotalCost + req.DeliveryCost + req.ServiceFee
	span.SetAttributes(attribute.Int64("order.total_cost", finalTotalCost))

	userDebts := make(map[int64]int64)

	if req.PayForAll {
		userDebts[req.UserID] = finalTotalCost
	} else {
		for _, item := range cart.Items {
			if item.OwnerUserID == nil {
				span.AddEvent("orphaned_items_found")
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
			split := domain.OrderSplit{
				ID:     uuid.New().String(),
				UserID: uid,
				Amount: amount,
				Status: SplitStatusPending,
			}
			// Если для основного плательщика выбрана сохранённая карта —
			// сохраняем yookassa-external-id, чтобы saga при CreatePayment
			// сразу списал с этой карты (а не открывал форму ввода).
			if uid == req.UserID && req.PaymentMethodID != "" {
				pm := req.PaymentMethodID
				split.PaymentMethodID = &pm
			}
			splits = append(splits, split)
		}
	}
	span.SetAttributes(attribute.Int("order.splits_count", len(splits)))

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
		return "", errutil.Internal("failed to save order to database", err)
	}
	span.SetAttributes(attribute.String("order.public_id", orderPublicID))

	cmd := events.SagaCommand{
		OrderID:        orderPublicID,
		UserID:         req.UserID,
		Action:         events.CommandLockCart,
		IdempotencyKey: idempotencyKey,
	}

	err = o.rabbitPublisher.PublishJSON(ctx, events.QueueCartCommands, cmd)
	if err != nil {
		span.AddEvent("saga_start_failed")
		return "", errutil.Internal("failed to start order saga", err)
	}

	span.AddEvent("saga_started")
	return orderPublicID, nil
}

// GetOrderPaymentID — возвращает yookassa_payment_id для конкретного заказа.
// Доступен только владельцу заказа (admin_account_id == userID).
func (o *orderUseCase) GetOrderPaymentID(ctx context.Context, orderPublicID string, userID int64) (string, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("order.public_id", orderPublicID),
		attribute.Int64("user.id", userID),
	)

	order, err := o.orderRepo.GetOrderByPublicID(ctx, orderPublicID)
	if err != nil {
		return "", errutil.Wrap("ORDER_NOT_FOUND", "order not found", err, codes.NotFound)
	}
	if order.AdminID != userID {
		return "", errutil.New("FORBIDDEN", "user is not the owner of this order", codes.PermissionDenied)
	}

	for _, sp := range order.Splits {
		if sp.YookassaPaymentID != nil && *sp.YookassaPaymentID != "" {
			return *sp.YookassaPaymentID, nil
		}
	}
	return "", errutil.New("PAYMENT_NOT_READY", "payment id not yet assigned to this order", codes.FailedPrecondition)
}

// CancelOrder помечает заказ как cancelled. Доступно владельцу и только пока
// заказ ещё не in_progress / delivering / finished (терминал).
func (o *orderUseCase) CancelOrder(ctx context.Context, orderPublicID string, userID int64) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("order.public_id", orderPublicID),
		attribute.Int64("user.id", userID),
	)

	order, err := o.orderRepo.GetOrderByPublicID(ctx, orderPublicID)
	if err != nil {
		return errutil.Wrap("ORDER_NOT_FOUND", "order not found", err, codes.NotFound)
	}
	if order.AdminID != userID {
		return errutil.New("FORBIDDEN", "user is not the owner of this order", codes.PermissionDenied)
	}

	switch order.Status {
	case StatusFinished, StatusCancelled:
		return errutil.New("ORDER_TERMINAL", "order already in terminal state", codes.FailedPrecondition)
	case StatusInProgress, StatusWaiting, StatusDelivering:
		return errutil.New("ORDER_IN_PROGRESS", "order is being prepared, cannot cancel", codes.FailedPrecondition)
	}

	if err := o.orderRepo.UpdateOrderStatus(ctx, orderPublicID, StatusCancelled); err != nil {
		return errutil.Internal("failed to cancel order", err)
	}
	return nil
}

func (o *orderUseCase) GetOrders(ctx context.Context, userID int64) ([]domain.Order, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int64("user.id", userID))

	orders, err := o.orderRepo.GetOrdersByUserID(ctx, userID)
	if err != nil {
		return []domain.Order{}, errutil.Internal("failed to get orders from repository", err)
	}

	if len(orders) == 0 {
		span.SetAttributes(attribute.Int("orders.count", 0))
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

	span.SetAttributes(
		attribute.Int("orders.count", len(orders)),
		attribute.Int("brands.unique_count", len(brandIDs)),
	)

	logos, err := o.restaurantClient.GetLogosByBrandIDs(ctx, brandIDs)
	if err != nil {
		span.AddEvent("restaurant_logos_fetch_failed")
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

func (o *orderUseCase) UpdateOrderStatusByPaymentID(ctx context.Context, paymentID string, status string, idempotencyKey string) (err error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("payment.id", paymentID),
		attribute.String("payment.status", status),
		attribute.String("idempotency_key", idempotencyKey),
	)

	if status != "succeeded" {
		span.AddEvent("status_not_succeeded_skipping")
		return nil
	}

	orderPublicID, err := o.orderRepo.UpdateSplitStatusByPaymentID(ctx, paymentID, SplitStatusPaid)
	if err != nil {
		return err
	}
	span.SetAttributes(attribute.String("order.public_id", orderPublicID))

	allPaid, err := o.orderRepo.AreAllSplitsPaid(ctx, orderPublicID)
	if err != nil {
		return err
	}
	span.SetAttributes(attribute.Bool("order.all_splits_paid", allPaid))

	if allPaid {
		err = o.orderRepo.UpdateOrderStatus(ctx, orderPublicID, StatusPaid)
		if err != nil {
			return err
		}

		// TODO: можно отправить событие в RestaurantService для начала готовки
		span.AddEvent("order_transition_to_paid")
		o.logger.Info("All splits paid!", logger.String("order_id", orderPublicID))

		gatewayEvent := events.GatewayEvent{
			OrderID: orderPublicID,
			Status:  StatusPaid,
		}
		_ = o.rabbitPublisher.PublishJSON(ctx, events.QueueGatewayEvents, gatewayEvent)
	}

	return nil
}

func (o *orderUseCase) PayForFriend(ctx context.Context, splitID string, adminID int64, paymentMethodID, idempotencyKey string) (err error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("split.id", splitID),
		attribute.Int64("admin.id", adminID),
		attribute.String("payment_method.id", paymentMethodID),
	)

	err = o.orderRepo.UpdateSplitPayer(ctx, splitID, adminID)
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

	span.AddEvent("publishing_payment_command")
	return o.rabbitPublisher.PublishJSON(ctx, events.QueuePaymentCommands, payCmd)
}

func (o *orderUseCase) ProcessSagaReply(ctx context.Context, reply events.SagaReply) (err error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("order.id", reply.OrderID),
		attribute.String("saga.step", reply.Step),
		attribute.String("saga.status", reply.Status),
	)

	if reply.Status == events.StatusError {
		span.SetAttributes(attribute.String("saga.error_details", reply.ErrorMessage))

		_ = o.orderRepo.UpdateOrderStatus(ctx, reply.OrderID, StatusFailed)
		if reply.SplitID != "" {
			_ = o.orderRepo.UpdateSplitStatus(ctx, reply.SplitID, SplitStatusFailed)
		}
		// Откат корзины
		if reply.Step == "PAYMENT" {
			span.AddEvent("compensating_cart_unlock")
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
		span.SetAttributes(attribute.Int("order.splits_count", len(order.Splits)))

		for _, split := range order.Splits {
			// Если split привязан к конкретной сохранённой карте — передаём
			// её external_id (YooKassa payment_method.id), чтобы YooKassa
			// сразу списала с этой карты, а не показывала форму ввода новой.
			pmID := ""
			if split.PaymentMethodID != nil {
				pmID = *split.PaymentMethodID
			}
			payCmd := events.SagaCommand{
				OrderID:         order.PublicID,
				SplitID:         split.ID,
				UserID:          split.UserID,
				Action:          events.CommandCreatePayment,
				Amount:          split.Amount,
				PaymentMethodID: pmID,
				IdempotencyKey:  split.ID + "_payment",
			}
			_ = o.rabbitPublisher.PublishJSON(ctx, events.QueuePaymentCommands, payCmd)
		}

	case "PAYMENT":
		span.SetAttributes(
			attribute.String("split.id", reply.SplitID),
			attribute.String("payment.external_id", reply.PaymentID),
		)

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
