package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/address/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/address/internal/repository"
)

//go:generate mockgen -destination=mocks/address_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/address AddressUseCase
type AddressUseCase interface {
	AddAddress(ctx context.Context, userID int64, addr domain.Address, idempotencyKey string) (string, error)
	GetMyAddresses(ctx context.Context, userID int64) ([]domain.Address, error)
	DeleteAddress(ctx context.Context, userID int64, addressPublicID string, idempotencyKey string) error
	UpdateAddress(ctx context.Context, userID int64, addr domain.Address, idempotencyKey string) error
}

type addressUseCase struct {
	repo repository.AddressRepository
}

func NewAddressUseCase(r repository.AddressRepository) AddressUseCase {
	return &addressUseCase{repo: r}
}

func (u *addressUseCase) AddAddress(ctx context.Context, userID int64, addr domain.Address, idempotencyKey string) (string, error) {
	// Тут можно добавить валидацию (например, координаты в пределах города)

	return u.repo.CreateAddress(ctx, userID, addr, idempotencyKey)
}

func (u *addressUseCase) GetMyAddresses(ctx context.Context, userID int64) ([]domain.Address, error) {
	return u.repo.GetAddressesByUserID(ctx, userID)
}

func (u *addressUseCase) DeleteAddress(ctx context.Context, userID int64, addressPublicID string, idempotencyKey string) error {
	return u.repo.DeleteAddress(ctx, userID, addressPublicID)
}

func (u *addressUseCase) UpdateAddress(ctx context.Context, userID int64, addr domain.Address, idempotencyKey string) error {
	return u.repo.UpdateAddress(ctx, userID, addr)
}
