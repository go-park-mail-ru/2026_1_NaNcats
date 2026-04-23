package usecase

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:generate mockgen -destination=mocks/cart_client_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/usecase CartClient
type CartClient interface {
	GetCart(ctx context.Context, userID int64) (domain.Cart, int64, error)
	ClearCart(ctx context.Context, userID int64, idempotencyKey string) error
}

//go:generate mockgen -destination=mocks/payment_client_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/usecase PaymentClient
type PaymentClient interface {
	CreatePayment(ctx context.Context, amount int64, paymentMethodID string, idempotencyKey string) (string, string, error)
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

//go:generate mockgen -destination=mocks/order_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/order OrderUseCase
type OrderUseCase interface {
	CreateOrder(ctx context.Context, userID int64, req domain.CreateOrderInput, idempotencyKey string) (string, string, error)
	GetOrders(ctx context.Context, userID int64) ([]domain.Order, error)
	UpdateOrderStatusByPaymentID(ctx context.Context, paymentID string, status string, idempotencyKey string) error
}

type orderUseCase struct {
	orderRepo                repository.OrderRepository
	addressClient            AddressClient
	cartClient               CartClient
	paymentClient            PaymentClient
	restaurantClient         RestaurantClient
	defaultRestaurantLogoURL string
}

func NewOrderUseCase(or repository.OrderRepository, ac AddressClient, cc CartClient, pc PaymentClient, rc RestaurantClient, drlurl string) OrderUseCase {
	return &orderUseCase{
		orderRepo:                or,
		addressClient:            ac,
		cartClient:               cc,
		paymentClient:            pc,
		restaurantClient:         rc,
		defaultRestaurantLogoURL: drlurl,
	}
}

func (o *orderUseCase) CreateOrder(ctx context.Context, userID int64, req domain.CreateOrderInput, idempotencyKey string) (string, string, error) {
	cart, cartTotalCost, err := o.cartClient.GetCart(ctx, userID)
	if err != nil {
		return "", "", errutil.Wrap("INTERNAL_SERVER_ERROR", "failed to get cart", err, codes.Internal)
	}
	if len(cart.Items) == 0 {
		return "", "", errutil.New("CART_EMPTY", "cart is empty", codes.InvalidArgument)
	}

	err = o.addressClient.CheckAddressExists(ctx, userID, req.AddressPublicID)
	if err != nil {
		st, ok := status.FromError(err)

		if ok && st.Code() == codes.NotFound {
			return "", "", errutil.New("ADDRESS_NOT_FOUND_OR_INVALID", "user provided bad address", codes.NotFound)
		}

		return "", "", fmt.Errorf("address service internal error: %w", err)
	}

	items := make([]domain.OrderDish, 0, len(cart.Items))
	for _, item := range cart.Items {
		items = append(items, domain.OrderDish{
			DishID:   item.DishID,
			Quantity: item.Quantity,
			Price:    item.Price,
		})
	}

	finalTotalCost := cartTotalCost + req.DeliveryCost + req.ServiceFee

	resName, err := o.restaurantClient.GetRestaurantName(ctx, req.RestaurantBranchID)
	if err != nil {
		return "", "", errutil.Wrap("INTERNAL_SERVER_ERROR", "failed to get restaurant info", err, codes.Internal)
	}

	order := domain.Order{
		ClientID:           userID,
		RestaurantBranchID: req.RestaurantBranchID,
		RestaurantBrandID:  req.RestaurantBrandID,
		RestaurantName:     resName,
		ClientAddressID:    req.AddressPublicID,
		TotalCost:          finalTotalCost,
		Status:             "waiting",
		Items:              items,
	}

	orderPublicID, err := o.orderRepo.CreateOrder(ctx, order, idempotencyKey)
	if err != nil {
		return "", "", errutil.Wrap("INTERNAL_SERVER_ERROR", "failed to create order in db", err, codes.Internal)
	}

	paymentID, confirmationURL, err := o.paymentClient.CreatePayment(ctx, finalTotalCost, req.PaymentMethodID, idempotencyKey)
	if err != nil {
		return "", "", errutil.Wrap("INTERNAL_SERVER_ERROR", "failed to initialize payment", err, codes.Internal)
	}

	if err = o.orderRepo.SetYookassaID(ctx, orderPublicID, paymentID); err != nil {
		return "", "", errutil.Wrap("INTERNAL_SERVER_ERROR", "failed to link yookassa ID to order", err, codes.Internal)
	}

	_ = o.cartClient.ClearCart(ctx, userID, idempotencyKey)

	return orderPublicID, confirmationURL, nil
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
		// печально, отдаем с дефолтным лого
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
