package usecase

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/api_clients/yookassa"
)

type OrderUseCase interface {
	CreateOrder(ctx context.Context, userID int, req domain.CreateOrderInput) (string, string, error)
	GetOrders(ctx context.Context, userID int) ([]domain.Order, error)
}

type orderUseCase struct {
	orderRepo      repository.OrderRepository
	addressRepo    repository.AddressRepository
	cartUC         CartUseCase
	yookassaClient *yookassa.Client
}

func NewOrderUseCase(or repository.OrderRepository, ar repository.AddressRepository, cuc CartUseCase, yc *yookassa.Client) *orderUseCase {
	return &orderUseCase{
		orderRepo:      or,
		addressRepo:    ar,
		cartUC:         cuc,
		yookassaClient: yc,
	}
}

func (o *orderUseCase) CreateOrder(ctx context.Context, userID int, req domain.CreateOrderInput) (string, string, error) {
	cart, totalCost, err := o.cartUC.GetCart(ctx, userID)
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

	order := domain.Order{
		ClientID:           userID,
		RestaurantBranchID: req.RestaurantBranchID,
		ClientAddressID:    clientAddressID,
		TotalCost:          totalCost,
		Status:             "in_progress",
		Items:              items,
	}

	orderPublicID, err := o.orderRepo.CreateOrder(ctx, order)
	if err != nil {
		return "", "", err
	}

	rubles := totalCost / 1_000_000
	kopecks := (totalCost%1_000_000)/10_000 + 100
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
		return "", "", nil
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
	return o.orderRepo.GetOrdersByUserID(ctx, userID)
}

