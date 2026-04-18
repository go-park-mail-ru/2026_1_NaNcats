package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/address/internal/domain"
)

type AddressRepository interface {
	CreateAddress(ctx context.Context, userID int64, addr domain.Address, idempotencyKey string) (string, error)
	GetAddressesByUserID(ctx context.Context, userID int64) ([]domain.Address, error)
	UpdateAddress(ctx context.Context, userID int64, addr domain.Address) error
	DeleteAddress(ctx context.Context, userID int64, publicID string) error
	GetInternalIDByPublicID(ctx context.Context, userID int64, publicID string) (int, error)
}
