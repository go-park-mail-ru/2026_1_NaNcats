package repository

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/domain"
)

var (
	ErrStateChanged  = errors.New("order state has changed or order not found")
	ErrSplitNotFound = errors.New("split not found by payment ID")
)

//go:generate mockgen -destination=mocks/order_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/repository OrderRepository

type OrderRepository interface {
	WithTransaction(ctx context.Context, fn func(txCtx context.Context) error) error

	CreateOrder(ctx context.Context, order domain.Order, idempotencyKey string) (int64, string, error)
	GetOrderByPublicID(ctx context.Context, publicID string) (domain.Order, error)
	GetOrdersByUserID(ctx context.Context, userID int64, limit, offset int32) ([]domain.Order, error)
	UpdateOrderStatus(ctx context.Context, publicID string, newStatus string, expectedStatuses ...string) error

	UpdateSplitStatusByPaymentID(ctx context.Context, yookassaPaymentID, newStatus string) (splitID string, orderPublicID string, err error)
	UpdateSplitStatus(ctx context.Context, splitID string, newStatus string) error
	AreAllSplitsPaid(ctx context.Context, orderPublicID string) (bool, error)
	SetSplitYookassaID(ctx context.Context, splitID string, yookassaID string) error
	UpdateSplitPayer(ctx context.Context, splitID string, newPayerID int64) error
	GetSplitByID(ctx context.Context, splitID string) (domain.OrderSplit, error)

	// Возвращает заказы, статусы которых нужно продвигать в фоне
	GetOrdersByStatuses(ctx context.Context, statuses []string) ([]domain.Order, error)

	// Бренды, у которых пользователь когда-либо оплачивал заказ (paid|finished).
	GetUserPaidBrands(ctx context.Context, userID int64) ([]int64, error)
	// Топ-N брендов по количеству оплат за окно windowDays.
	GetTrendingBrands(ctx context.Context, windowDays, limit int32) ([]int64, error)
	// Топ-N блюд ресторана по SUM(quantity) за окно windowDays, paid|finished.
	GetTopDishesByBrand(ctx context.Context, brandID int64, windowDays, limit int32) ([]int64, error)

	GetPromocodeByCodeWithLock(ctx context.Context, code string) (domain.Promocode, error)
	CheckPromocodeUsage(ctx context.Context, promoID, userID int64) (bool, error)
	IncrementPromocodeUses(ctx context.Context, promoID int64) error
	CreatePromocodeUsage(ctx context.Context, promoID, orderID, userID int64) error

	RollbackPromocodeUsage(ctx context.Context, orderPublicID string) error
}
