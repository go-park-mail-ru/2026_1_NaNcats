package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
)

type AddressUseCase interface {
	AddAddress(ctx context.Context, userID int, addr domain.Address) (string, error)
	GetMyAddresses(ctx context.Context, userID int) ([]domain.Address, error)
	DeleteAddress(ctx context.Context, userID int, addressPublicID string) error
}

type addressUseCase struct {
	repo repository.AddressRepository
}

func NewAddressUseCase(r repository.AddressRepository) AddressUseCase {
	return &addressUseCase{repo: r}
}

func (u *addressUseCase) AddAddress(ctx context.Context, userID int, addr domain.Address) (string, error) {
	// Тут можно добавить валидацию (например, координаты в пределах города)
	return u.repo.CreateAddress(ctx, userID, addr)
}

func (u *addressUseCase) GetMyAddresses(ctx context.Context, userID int) ([]domain.Address, error) {
	return u.repo.GetAddressesByUserID(ctx, userID)
}

func (u *addressUseCase) DeleteAddress(ctx context.Context, userID int, addressPublicID string) error {
	return u.repo.DeleteAddress(ctx, userID, addressPublicID)
}
