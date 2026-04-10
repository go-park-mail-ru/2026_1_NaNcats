package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/api_clients/yookassa"
)

type OrderUseCase interface {
	CreateOrder(ctx context.Context, userID int, req domain.CreateOrderInput) (string, string, error)
	ProcessPaymentWebhook(ctx context.Context, notification *yookassa.WebhookNotification) error
}

type orderUseCase struct {
	orderRepo      repository.OrderRepository
	addressRepo    repository.AddressRepository
	cartUC         CartUseCase
	yookassaClient *yookassa.Client
}

func NewOrderUseCase(or repository.OrderRepository) *orderUseCase {
	return &orderUseCase{
		orderRepo: or,
	}
}

func (o *orderUseCase) CreateOrder(ctx context.Context, userID int, req domain.CreateOrderInput) (string, string, error) {
	cart, totalCost, err := o.cartUC.GetCart(ctx, userID)
	if err != nil {
		return "", "", err
	}

	_, _ = cart, totalCost

	return "", "", nil
}
