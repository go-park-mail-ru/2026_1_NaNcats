package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
)

type OrderRepository interface {
	CreateOrder(ctx context.Context, order domain.Order) (string, error)
	UpdateStatusByPaymentID(ctx context.Context, yookassaPaymentID, newStatus string) error
	GetOrderByPublicID(ctx context.Context, publicID string, userID int) (domain.Order, error)
	SetYookassaID(ctx context.Context, orderPublicID, yookassaID string) error
	GetOrdersByUserID(ctx context.Context, userID int) ([]domain.Order, error)
}
