package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/domain"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order domain.Order, idempotencyKey string) (string, error)
	GetOrderByPublicID(ctx context.Context, publicID string) (domain.Order, error)
	GetOrdersByUserID(ctx context.Context, userID int64) ([]domain.Order, error)
	UpdateOrderStatus(ctx context.Context, publicID string, newStatus string) error

	UpdateSplitStatusByPaymentID(ctx context.Context, yookassaPaymentID, newStatus string) (string, error)
	UpdateSplitStatus(ctx context.Context, splitID string, newStatus string) error
	AreAllSplitsPaid(ctx context.Context, orderPublicID string) (bool, error)
	SetSplitYookassaID(ctx context.Context, splitID string, yookassaID string) error
	UpdateSplitPayer(ctx context.Context, splitID string, newPayerID int64) error
	GetSplitByID(ctx context.Context, splitID string) (domain.OrderSplit, error)

	// Возвращает заказы, статусы которых нужно продвигать в фоне.
	GetOrdersByStatuses(ctx context.Context, statuses []string) ([]domain.Order, error)
}
