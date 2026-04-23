package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/domain"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order domain.Order, idempotencyKey string) (string, error)
	UpdateStatusByPaymentID(ctx context.Context, yookassaPaymentID, newStatus string) error
	UpdateOrderStatus(ctx context.Context, publicID string, newStatus string) error
	GetOrderByPublicID(ctx context.Context, publicID string, userID int64) (domain.Order, error)
	SetYookassaID(ctx context.Context, orderPublicID, yookassaID string) error
	GetOrdersByUserID(ctx context.Context, userID int64) ([]domain.Order, error)
}
