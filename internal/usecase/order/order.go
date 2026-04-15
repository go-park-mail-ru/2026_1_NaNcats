package usecase

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
	cart "github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/cart"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/api_clients/yookassa"
)

//go:generate mockgen -destination=mocks/order_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase OrderUseCase
type OrderUseCase interface {
	CreateOrder(ctx context.Context, userID int, req domain.CreateOrderInput) (string, string, error)
	GetOrders(ctx context.Context, userID int) ([]domain.Order, error)
}

type orderUseCase struct {
	orderRepo                repository.OrderRepository
	addressRepo              repository.AddressRepository
	cartUC                   cart.CartUseCase
	yookassaClient           *yookassa.Client
	defaultRestaurantLogoURL string
}

func NewOrderUseCase(or repository.OrderRepository, ar repository.AddressRepository, cuc cart.CartUseCase, yc *yookassa.Client, drlurl string) *orderUseCase {
	return &orderUseCase{
		orderRepo:                or,
		addressRepo:              ar,
		cartUC:                   cuc,
		yookassaClient:           yc,
		defaultRestaurantLogoURL: drlurl,
	}
}

func (o *orderUseCase) CreateOrder(ctx context.Context, userID int, req domain.CreateOrderInput) (string, string, error) {
	// 1. Получаем стоимость ТОЛЬКО товаров в корзине
	cart, cartTotalCost, err := o.cartUC.GetCart(ctx, userID)
	if err != nil {
		return "", "", err
	}

	clientAddressID, err := o.addressRepo.GetInternalIDByPublicID(ctx, userID, req.AddressPublicID)
	if err != nil {
		return "", "", domain.ErrAddressNotFound
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

	order := domain.Order{
		ClientID:           userID,
		RestaurantBranchID: req.RestaurantBranchID,
		ClientAddressID:    clientAddressID,
		TotalCost:          finalTotalCost,
		Status:             "in_progress",
		Items:              items,
	}

	orderPublicID, err := o.orderRepo.CreateOrder(ctx, order)
	if err != nil {
		return "", "", err
	}

	rubles := finalTotalCost / 1_000_000
	kopecks := (finalTotalCost%1_000_000)/10_000 + 100
	value := strconv.FormatInt(rubles, 10) + "." + strconv.FormatInt(kopecks, 10)[1:]

	paymentRequest := yookassa.CreatePaymentRequest{
		Amount: yookassa.CreatePaymentRequestAmount{
			Value:    value,
			Currency: "RUB",
		},
		Capture:           true,
		SavePaymentMethod: false,
	}

	paymentRequest.Confirmation = &yookassa.CreatePaymentRequestConfirmation{
		Type:      "redirect",
		ReturnURL: os.Getenv("YOOKASSA_RETURN_URL"),
	}

	if req.PaymentMethodID != "" {
		paymentRequest.PaymentMethodID = req.PaymentMethodID
	}

	paymentResponse, err := o.yookassaClient.CreatePayment(ctx, paymentRequest)
	if err != nil {
		return "", "", err
	}

	if err = o.orderRepo.SetYookassaID(ctx, orderPublicID, paymentResponse.ID); err != nil {
		return "", "", fmt.Errorf("failed to link yookassa ID: %w", err)
	}
	_ = o.cartUC.UpdateCart(ctx, userID, domain.Cart{})

	var confirmationURL string
	if paymentResponse.Confirmation != nil && paymentResponse.Confirmation.Type == "redirect" {
		confirmationURL = paymentResponse.Confirmation.ConfirmationURL
	}

	return orderPublicID, confirmationURL, nil
}

func (o *orderUseCase) GetOrders(ctx context.Context, userID int) ([]domain.Order, error) {
	orders, err := o.orderRepo.GetOrdersByUserID(ctx, userID)
	if err != nil {
		return []domain.Order{}, err
	}

	for i, order := range orders {
		if order.RestaurantLogoURL == "" {
			orders[i].RestaurantLogoURL = o.defaultRestaurantLogoURL
		}
	}

	return orders, nil
}
