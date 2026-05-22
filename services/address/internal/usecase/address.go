package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/address/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/address/internal/repository"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

//go:generate gowrap gen -i AddressUseCase -t ../../../../shared/templates/tracing.tmpl -o address_tracing_mw.go -v TracerName=address-service
//go:generate mockgen -destination=mocks/address_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/address/internal/usecase AddressUseCase
type AddressUseCase interface {
	AddAddress(ctx context.Context, userID int64, addr domain.Address, idempotencyKey string) (string, error)
	GetMyAddresses(ctx context.Context, userID int64) ([]domain.Address, error)
	DeleteAddress(ctx context.Context, userID int64, addressPublicID string, idempotencyKey string) error
	UpdateAddress(ctx context.Context, userID int64, addr domain.Address, idempotencyKey string) error
	CheckAddressExists(ctx context.Context, userID int64, addressPublicID string) error
}

type addressUseCase struct {
	repo repository.AddressRepository
}

func NewAddressUseCase(r repository.AddressRepository) AddressUseCase {
	return &addressUseCase{repo: r}
}

func (u *addressUseCase) AddAddress(ctx context.Context, userID int64, addr domain.Address, idempotencyKey string) (string, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("user.id", userID),
		attribute.String("address.label", addr.Label),
		attribute.String("idempotency.key", idempotencyKey),
	)

	return u.repo.CreateAddress(ctx, userID, addr, idempotencyKey)
}

func (u *addressUseCase) GetMyAddresses(ctx context.Context, userID int64) ([]domain.Address, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int64("user.id", userID))

	addresses, err := u.repo.GetAddressesByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	span.SetAttributes(attribute.Int("addresses.count", len(addresses)))
	return addresses, nil
}

func (u *addressUseCase) DeleteAddress(ctx context.Context, userID int64, addressPublicID string, idempotencyKey string) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("user.id", userID),
		attribute.String("address.public_id", addressPublicID),
		attribute.String("idempotency.key", idempotencyKey),
	)

	return u.repo.DeleteAddress(ctx, userID, addressPublicID, idempotencyKey)
}

func (u *addressUseCase) UpdateAddress(ctx context.Context, userID int64, addr domain.Address, idempotencyKey string) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("user.id", userID),
		attribute.String("address.public_id", addr.PublicID),
		attribute.String("idempotency.key", idempotencyKey),
	)

	return u.repo.UpdateAddress(ctx, userID, addr, idempotencyKey)
}

func (u *addressUseCase) CheckAddressExists(ctx context.Context, userID int64, addressPublicID string) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("user.id", userID),
		attribute.String("address.public_id", addressPublicID),
	)

	return u.repo.CheckAddressExists(ctx, userID, addressPublicID)
}
