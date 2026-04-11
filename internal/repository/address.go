package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
)

type AddressRepository interface {
	CreateAddress(ctx context.Context, userID int, addr domain.Address) (string, error)
	GetAddressesByUserID(ctx context.Context, userID int) ([]domain.Address, error)
	UpdateAddress(ctx context.Context, userID int, addr domain.Address) error
	DeleteAddress(ctx context.Context, userID int, publicID string) error
	GetInternalIDByPublicID(ctx context.Context, userID int, publicID string) (int, error)
}
