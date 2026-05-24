package usecase

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

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

	// StatusSplitPaid не является статусом заказа: это сигнал фронту об
	// оплате одной доли счёта, по нему модалка помечает долю оплаченной.
	StatusSplitPaid = "split_paid"
)

//go:generate mockgen -destination=mocks/order_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/usecase OrderUseCase,CartClient,AddressClient,RestaurantClient,UserClient,MessagePublisher
//go:generate gowrap gen -i OrderUseCase -t ../../../../shared/templates/tracing.tmpl -o order_tracing_mw.go -v TracerName=order-service

type CartClient interface {
	GetCart(ctx context.Context, userID int64) (domain.Cart, int64, error)
	LockCart(ctx context.Context, cartID string, userID int64, idempotencyKey string) error
	UnlockCart(ctx context.Context, cartID string, userID int64, idempotencyKey string) error
}

type AddressClient interface {
	CheckAddressExists(ctx context.Context, userID int64, addressPublicID string) error
}

type RestaurantClient interface {
	GetRestaurantName(ctx context.Context, branchID int64) (string, error)
}

type UserClient interface {
	OnOrderPaid(ctx context.Context, userID, restaurantID int64, orderPublicID string, paidAt time.Time) error
}

type MessagePublisher interface {
	PublishJSON(ctx context.Context, queueName string, data any) error
}

type OrderUseCase interface {
	CreateOrder(ctx context.Context, req domain.CreateOrderInput, idempotencyKey string) (string, error)
	GetOrders(ctx context.Context, userID int64, limit, offset int32) ([]domain.Order, error)
	UpdateOrderStatusByPaymentID(ctx context.Context, paymentID string, status string, idempotencyKey string) error
	ProcessSagaReply(ctx context.Context, reply events.SagaReply) error
	PayForFriend(ctx context.Context, splitID string, adminID int64, paymentMethodID, idempotencyKey string) error
	CancelOrder(ctx context.Context, orderPublicID string, userID int64) error
	GetUserPaidBrands(ctx context.Context, userID int64) ([]int64, error)
	GetTrendingBrands(ctx context.Context, windowDays, limit int32) ([]int64, error)
	GetTopDishesByBrand(ctx context.Context, brandID int64, windowDays, limit int32) ([]int64, error)
}

type orderUseCase struct {
	orderRepo                repository.OrderRepository
	addressClient            AddressClient
	cartClient               CartClient
	restaurantClient         RestaurantClient
	userClient               UserClient
	rabbitPublisher          MessagePublisher
	defaultRestaurantLogoURL string
	logger                   logger.Logger
}

func NewOrderUseCase(
	or repository.OrderRepository,
	ac AddressClient,
	cc CartClient,
	rc RestaurantClient,
	uc UserClient,
	rp MessagePublisher,
	drlurl string,
	l logger.Logger,
) OrderUseCase {
	return &orderUseCase{
		orderRepo:                or,
		addressClient:            ac,
		cartClient:               cc,
		restaurantClient:         rc,
		userClient:               uc,
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

	lockErr := o.cartClient.LockCart(ctx, cart.ID, req.UserID, idempotencyKey+"_lock")
	if lockErr != nil {
		return "", errutil.Wrap("CART_LOCKED", "failed to lock cart or unassigned items exist", lockErr, codes.FailedPrecondition)
	}

	var orderPublicID string
	var transactionErr error
	var createdSplits []domain.OrderSplit

	defer func() {
		if transactionErr != nil {
			span.AddEvent("rollback_cart_lock")
			_ = o.cartClient.UnlockCart(context.Background(), cart.ID, req.UserID, idempotencyKey+"_unlock")
		}
	}()

	transactionErr = o.orderRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		var appliedPromo *domain.Promocode
		var totalDiscount int64 = 0

		if req.Promocode != nil && *req.Promocode != "" {
			promo, promoErr := o.orderRepo.GetPromocodeByCodeWithLock(txCtx, *req.Promocode)
			if promoErr != nil {
				return errutil.Wrap("PROMO_NOT_FOUND", "promocode not found or inactive", promoErr, codes.NotFound)
			}

			if time.Now().After(promo.ExpiresAt) {
				return errutil.New("PROMO_EXPIRED", "promocode has expired", codes.FailedPrecondition)
			}
			if promo.MaxUses != nil && promo.CurrentUses >= *promo.MaxUses {
				return errutil.New("PROMO_LIMIT_REACHED", "promocode usage limit reached", codes.FailedPrecondition)
			}
			if promo.MinOrderAmount != nil && cartTotalCost < *promo.MinOrderAmount {
				return errutil.New("PROMO_MIN_AMOUNT", "cart total is less than minimum order amount", codes.FailedPrecondition)
			}
			if promo.RestaurantBrandID != nil && *promo.RestaurantBrandID != req.RestaurantBrandID {
				return errutil.New("PROMO_INVALID_RESTAURANT", "promocode is not valid for this restaurant", codes.FailedPrecondition)
			}
			if promo.UserID != nil && *promo.UserID != req.UserID {
				return errutil.New("PROMO_FORBIDDEN", "promocode is tied to another user", codes.PermissionDenied)
			}

			used, checkErr := o.orderRepo.CheckPromocodeUsage(txCtx, promo.ID, req.UserID)
			if checkErr != nil {
				return errutil.Internal("failed to check promo usage", checkErr)
			}
			if used {
				return errutil.New("PROMO_ALREADY_USED", "you have already used this promocode", codes.FailedPrecondition)
			}

			if promo.DiscountAmount != nil {
				totalDiscount = *promo.DiscountAmount
				if totalDiscount > cartTotalCost {
					totalDiscount = cartTotalCost
				}
			} else if promo.DiscountPercent != nil {
				totalDiscount = (cartTotalCost * int64(*promo.DiscountPercent)) / 100
			}

			appliedPromo = &promo
		}

		finalTotalCost := cartTotalCost - totalDiscount + req.DeliveryCost + req.ServiceFee
		span.SetAttributes(attribute.Int64("order.final_total_cost", finalTotalCost))

		userDebts := make(map[int64]int64)
		userDiscounts := make(map[int64]int64)

		if req.PayForAll {
			userDebts[req.UserID] = cartTotalCost
			userDiscounts[req.UserID] = totalDiscount
		} else {
			for _, item := range cart.Items {
				userDebts[*item.OwnerUserID] += item.Price * int64(item.Quantity)
			}

			for targetID, payerID := range req.PayerMapping {
				if debt, exists := userDebts[targetID]; exists {
					userDebts[payerID] += debt
					delete(userDebts, targetID)
				}
			}

			var assignedDiscount int64 = 0
			cartTotalBig := big.NewInt(cartTotalCost)
			totalDiscBig := big.NewInt(totalDiscount)

			for uid, debt := range userDebts {
				if debt > 0 {
					debtBig := big.NewInt(debt)
					rawDiscountBig := new(big.Int).Mul(debtBig, totalDiscBig)
					rawDiscountBig.Div(rawDiscountBig, cartTotalBig)
					rawDiscount := rawDiscountBig.Int64()
					userDiscount := (rawDiscount / 10000) * 10000

					userDiscounts[uid] = userDiscount
					assignedDiscount += userDiscount
				}
			}

			remainder := totalDiscount - assignedDiscount
			if remainder > 0 {
				userDiscounts[req.UserID] += remainder
			}
		}

		splits := make([]domain.OrderSplit, 0, len(userDebts))
		for uid, foodDebt := range userDebts {
			if foodDebt > 0 {
				userDisc := userDiscounts[uid]
				finalAmount := foodDebt - userDisc
				if uid == req.UserID {
					finalAmount += req.DeliveryCost + req.ServiceFee
				}

				split := domain.OrderSplit{
					ID:             uuid.New().String(),
					UserID:         uid,
					BaseAmount:     foodDebt,
					DiscountAmount: userDisc,
					Amount:         finalAmount,
					Status:         SplitStatusPending,
				}
				if uid == req.UserID && req.PaymentMethodID != "" {
					pm := req.PaymentMethodID
					split.PaymentMethodID = &pm
				}
				splits = append(splits, split)
			}
		}

		items := make([]domain.OrderDish, 0, len(cart.Items))
		for _, item := range cart.Items {
			//nolint:staticcheck
			items = append(items, domain.OrderDish{
				DishID:      item.DishID,
				Name:        item.Name,
				Quantity:    item.Quantity,
				Price:       item.Price,
				OwnerUserID: item.OwnerUserID,
			})
		}

		var promoIDPtr *int64
		var promoStrPtr *string
		if appliedPromo != nil {
			promoIDPtr = &appliedPromo.ID
			promoStrPtr = &appliedPromo.Code
		}

		order := domain.Order{
			AdminID:            req.UserID,
			RestaurantBranchID: req.RestaurantBranchID,
			RestaurantBrandID:  req.RestaurantBrandID,
			RestaurantName:     resName,
			ClientAddressID:    req.AddressPublicID,
			TotalCost:          finalTotalCost,
			Status:             StatusCartLocked,
			PromocodeID:        promoIDPtr,
			PromocodeString:    promoStrPtr,
			DiscountAmount:     totalDiscount,
			Items:              items,
			Splits:             splits,
		}

		internalID, pubID, repoErr := o.orderRepo.CreateOrder(txCtx, order, idempotencyKey)
		if repoErr != nil {
			return errutil.Internal("failed to save order to database", repoErr)
		}

		orderPublicID = pubID
		createdSplits = splits

		if internalID != 0 && appliedPromo != nil {
			if err := o.orderRepo.IncrementPromocodeUses(txCtx, appliedPromo.ID); err != nil {
				return errutil.Internal("failed to increment promo uses", err)
			}
			if err := o.orderRepo.CreatePromocodeUsage(txCtx, appliedPromo.ID, internalID, req.UserID); err != nil {
				return errutil.Internal("failed to record promo usage", err)
			}
		}

		return nil
	})

	if transactionErr != nil {
		span.RecordError(transactionErr)
		return "", transactionErr
	}

	span.SetAttributes(attribute.String("order.public_id", orderPublicID))

	for _, split := range createdSplits {
		pmID := ""
		if split.PaymentMethodID != nil {
			pmID = *split.PaymentMethodID
		}
		payCmd := events.SagaCommand{
			OrderID:         orderPublicID,
			SplitID:         split.ID,
			UserID:          split.UserID,
			Action:          events.CommandCreatePayment,
			Amount:          split.Amount,
			PaymentMethodID: pmID,
			IdempotencyKey:  split.ID + "_payment",
		}
		_ = o.rabbitPublisher.PublishJSON(ctx, events.QueuePaymentCommands, payCmd)
	}

	span.AddEvent("saga_started_directly_to_payment")
	return orderPublicID, nil
}

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

	err = o.orderRepo.UpdateOrderStatus(ctx, orderPublicID, StatusCancelled, StatusCreated, StatusCartLocked, StatusPaymentReady, StatusPaid)
	if err != nil {
		if errors.Is(err, repository.ErrStateChanged) {
			return errutil.New("ORDER_IN_PROGRESS_OR_TERMINAL", "order is being prepared or already finished, cannot cancel", codes.FailedPrecondition)
		}
		return errutil.Internal("failed to cancel order", err)
	}

	if err := o.orderRepo.RollbackPromocodeUsage(ctx, orderPublicID); err != nil {
		span.RecordError(err)
		o.logger.Error("failed to rollback promocode on manual cancel", err, logger.String("order_id", orderPublicID))
		return errutil.Internal("order cancelled, but failed to rollback promocode", err)
	}

	span.AddEvent("order_cancelled_and_promocode_rolled_back")
	return nil
}

func (o *orderUseCase) GetOrders(ctx context.Context, userID int64, limit, offset int32) ([]domain.Order, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("user.id", userID),
		attribute.Int("query.limit", int(limit)),
		attribute.Int("query.offset", int(offset)),
	)

	orders, err := o.orderRepo.GetOrdersByUserID(ctx, userID, limit, offset)
	if err != nil {
		return []domain.Order{}, errutil.Internal("failed to get orders from repository", err)
	}

	span.SetAttributes(attribute.Int("orders.count", len(orders)))

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

	splitID, orderPublicID, err := o.orderRepo.UpdateSplitStatusByPaymentID(ctx, paymentID, SplitStatusPaid)
	if err != nil {
		if errors.Is(err, repository.ErrSplitNotFound) {
			return errutil.Wrap("SPLIT_NOT_FOUND", "split not found by payment ID", err, codes.NotFound)
		}
		return err
	}
	span.SetAttributes(
		attribute.String("order.public_id", orderPublicID),
		attribute.String("split.id", splitID),
	)

	// Отдельным событием сообщаем фронту, что эта доля счёта оплачена.
	_ = o.rabbitPublisher.PublishJSON(ctx, events.QueueGatewayEvents, events.GatewayEvent{
		OrderID: orderPublicID,
		SplitID: splitID,
		Status:  StatusSplitPaid,
	})

	allPaid, err := o.orderRepo.AreAllSplitsPaid(ctx, orderPublicID)
	if err != nil {
		return err
	}
	span.SetAttributes(attribute.Bool("order.all_splits_paid", allPaid))

	if allPaid {
		err = o.orderRepo.UpdateOrderStatus(ctx, orderPublicID, StatusPaid, StatusCreated, StatusCartLocked, StatusPaymentReady)
		if err != nil {
			if errors.Is(err, repository.ErrStateChanged) {
				span.AddEvent("order_already_advanced_ignoring_paid")
				return nil
			}
			return err
		}

		span.AddEvent("order_transition_to_paid")
		o.logger.Info("All splits paid!", logger.String("order_id", orderPublicID))

		gatewayEvent := events.GatewayEvent{
			OrderID: orderPublicID,
			Status:  StatusPaid,
		}
		_ = o.rabbitPublisher.PublishJSON(ctx, events.QueueGatewayEvents, gatewayEvent)

		paidOrder, getErr := o.orderRepo.GetOrderByPublicID(ctx, orderPublicID)
		if getErr != nil {
			o.logger.WithContext(ctx).Warn("failed to load order for OnOrderPaid hook",
				logger.String("order_id", orderPublicID), logger.Err(getErr))
		} else if o.userClient != nil {
			if hookErr := o.userClient.OnOrderPaid(ctx, paidOrder.AdminID, paidOrder.RestaurantBranchID, orderPublicID, time.Now()); hookErr != nil {
				o.logger.WithContext(ctx).Warn("user-service OnOrderPaid hook failed",
					logger.String("order_id", orderPublicID), logger.Err(hookErr))
			}
		}
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

	split, err := o.orderRepo.GetSplitByID(ctx, splitID)
	if err != nil {
		return errutil.Wrap("SPLIT_NOT_FOUND", "split not found", err, codes.NotFound)
	}

	if split.Status != SplitStatusPending {
		return errutil.New("SPLIT_NOT_PENDING", "split is already paid, failed, or cancelled", codes.FailedPrecondition)
	}

	err = o.orderRepo.UpdateSplitPayer(ctx, splitID, adminID)
	if err != nil {
		return errutil.Wrap("CANNOT_REASSIGN", "failed to reassign split", err, codes.InvalidArgument)
	}

	// В сагу передаём публичный UUID заказа: по нему gateway находит нужный
	// WebSocket-канал. С внутренним числовым id событие оплаты не дойдёт.
	payCmd := events.SagaCommand{
		OrderID:         split.OrderPublicID,
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
		o.logger.Error("saga step failed", fmt.Errorf("%s", reply.ErrorMessage),
			logger.String("order_id", reply.OrderID),
			logger.String("step", reply.Step),
			logger.String("split_id", reply.SplitID),
		)

		_ = o.orderRepo.UpdateOrderStatus(ctx, reply.OrderID, StatusFailed, StatusCreated, StatusCartLocked, StatusPaymentReady)

		if reply.SplitID != "" {
			_ = o.orderRepo.UpdateSplitStatus(ctx, reply.SplitID, SplitStatusFailed)
		}

		if err := o.orderRepo.RollbackPromocodeUsage(ctx, reply.OrderID); err != nil {
			span.RecordError(err)
			o.logger.Error("failed to rollback promocode in saga", err, logger.String("order_id", reply.OrderID))
			return err
		}

		if reply.Step == "PAYMENT" {
			// Платёж не создался, но сам заказ валиден: переводим в
			// payment_ready и шлём событие с ошибкой, чтобы фронт предложил
			// повторить оплату или сменить карту, а не убивал заказ.
			_ = o.orderRepo.UpdateOrderStatus(ctx, reply.OrderID, StatusPaymentReady)
			if reply.SplitID != "" {
				_ = o.orderRepo.UpdateSplitStatus(ctx, reply.SplitID, SplitStatusFailed)
			}

			gatewayEvent := events.GatewayEvent{
				OrderID: reply.OrderID,
				SplitID: reply.SplitID,
				UserID:  reply.UserID,
				Status:  StatusPaymentReady,
				Error:   reply.ErrorMessage,
			}
			_ = o.rabbitPublisher.PublishJSON(ctx, events.QueueGatewayEvents, gatewayEvent)
		} else {
			// Сбой не на шаге оплаты — заказ помечаем проваленным.
			_ = o.orderRepo.UpdateOrderStatus(ctx, reply.OrderID, StatusFailed)
			if reply.SplitID != "" {
				_ = o.orderRepo.UpdateSplitStatus(ctx, reply.SplitID, SplitStatusFailed)
			}
		}
		return nil
	}

	switch reply.Step {
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

func (o *orderUseCase) GetUserPaidBrands(ctx context.Context, userID int64) ([]int64, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int64("user.id", userID))

	brands, err := o.orderRepo.GetUserPaidBrands(ctx, userID)
	if err != nil {
		return nil, errutil.Internal("failed to fetch user paid brands", err)
	}
	span.SetAttributes(attribute.Int("user.paid_brands.count", len(brands)))
	return brands, nil
}

func (o *orderUseCase) GetTrendingBrands(ctx context.Context, windowDays, limit int32) ([]int64, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int("trending.window_days", int(windowDays)),
		attribute.Int("trending.limit", int(limit)),
	)

	brands, err := o.orderRepo.GetTrendingBrands(ctx, windowDays, limit)
	if err != nil {
		return nil, errutil.Internal("failed to fetch trending brands", err)
	}
	span.SetAttributes(attribute.Int("trending.brands.count", len(brands)))
	return brands, nil
}

func (o *orderUseCase) GetTopDishesByBrand(ctx context.Context, brandID int64, windowDays, limit int32) ([]int64, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("brand.id", brandID),
		attribute.Int("top_dishes.window_days", int(windowDays)),
		attribute.Int("top_dishes.limit", int(limit)),
	)

	ids, err := o.orderRepo.GetTopDishesByBrand(ctx, brandID, windowDays, limit)
	if err != nil {
		return nil, errutil.Internal("failed to fetch top dishes by brand", err)
	}
	span.SetAttributes(attribute.Int("top_dishes.count", len(ids)))
	return ids, nil
}
