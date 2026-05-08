package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/address/internal/domain"
)

//go:generate mockgen -destination=mocks/address_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/address/internal/repository AddressRepository
type AddressRepository interface {
	CreateAddress(ctx context.Context, userID int64, addr domain.Address, idempotencyKey string) (string, error)
	GetAddressesByUserID(ctx context.Context, userID int64) ([]domain.Address, error)
	UpdateAddress(ctx context.Context, userID int64, addr domain.Address, idempotencyKey string) error
	DeleteAddress(ctx context.Context, userID int64, publicID string, idempotencyKey string) error
	GetInternalIDByPublicID(ctx context.Context, userID int64, publicID string) (int, error)
	CheckAddressExists(ctx context.Context, userID int64, publicID string) error
}
