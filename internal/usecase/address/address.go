package usecase

import (
	"context"
	"html"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
)

//go:generate mockgen -destination=mocks/address_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase AddressUseCase
type AddressUseCase interface {
	AddAddress(ctx context.Context, userID int, addr domain.Address) (string, error)
	GetMyAddresses(ctx context.Context, userID int) ([]domain.Address, error)
	DeleteAddress(ctx context.Context, userID int, addressPublicID string) error
	UpdateAddress(ctx context.Context, userID int, addr domain.Address) error
}

type addressUseCase struct {
	repo repository.AddressRepository
}

func NewAddressUseCase(r repository.AddressRepository) AddressUseCase {
	return &addressUseCase{repo: r}
}

func (u *addressUseCase) AddAddress(ctx context.Context, userID int, addr domain.Address) (string, error) {
	createAddr := domain.Address{
		PublicID: addr.PublicID,
		Location: domain.Location{
			AddressText: html.EscapeString(addr.Location.AddressText),
			Latitude:    addr.Location.Latitude,
			Longitude:   addr.Location.Longitude,
		},
		Apartment:      html.EscapeString(addr.Apartment),
		Entrance:       html.EscapeString(addr.Entrance),
		Floor:          html.EscapeString(addr.Floor),
		DoorCode:       html.EscapeString(addr.DoorCode),
		CourierComment: html.EscapeString(addr.CourierComment),
		Label:          html.EscapeString(addr.Label),
	}
	// Тут можно добавить валидацию (например, координаты в пределах города)

	return u.repo.CreateAddress(ctx, userID, createAddr)
}

func (u *addressUseCase) GetMyAddresses(ctx context.Context, userID int) ([]domain.Address, error) {
	return u.repo.GetAddressesByUserID(ctx, userID)
}

func (u *addressUseCase) DeleteAddress(ctx context.Context, userID int, addressPublicID string) error {
	return u.repo.DeleteAddress(ctx, userID, addressPublicID)
}

func (u *addressUseCase) UpdateAddress(ctx context.Context, userID int, addr domain.Address) error {
	updateAddr := domain.Address{
		PublicID: addr.PublicID,
		Location: domain.Location{
			AddressText: html.EscapeString(addr.Location.AddressText),
			Latitude:    addr.Location.Latitude,
			Longitude:   addr.Location.Longitude,
		},
		Apartment:      html.EscapeString(addr.Apartment),
		Entrance:       html.EscapeString(addr.Entrance),
		Floor:          html.EscapeString(addr.Floor),
		DoorCode:       html.EscapeString(addr.DoorCode),
		CourierComment: html.EscapeString(addr.CourierComment),
		Label:          html.EscapeString(addr.Label),
	}

	return u.repo.UpdateAddress(ctx, userID, updateAddr)
}
