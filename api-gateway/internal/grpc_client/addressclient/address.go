package addressclient

import (
	"context"
	"errors"

	pbAddress "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/address"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrAddressNotFound = errors.New("address not found")
	ErrInternal        = errors.New("internal server error")
)

type Location struct {
	AddressText string
	Latitude    float64
	Longitude   float64
}

type Address struct {
	PublicID       string
	Location       Location
	Apartment      string
	Entrance       string
	Floor          string
	DoorCode       string
	CourierComment string
	Label          string
}

type AddressClient interface {
	AddAddress(ctx context.Context, userID int64, addr Address, idempotencyKey string) (string, error)
	GetMyAddresses(ctx context.Context, userID int64) ([]Address, error)
	DeleteAddress(ctx context.Context, userID int64, publicID string, idempotencyKey string) error
	UpdateAddress(ctx context.Context, userID int64, addr Address, idempotencyKey string) error
	CheckAddressExists(ctx context.Context, userID int64, publicID string) error
}

type addressClient struct {
	client pbAddress.AddressServiceClient
}

func NewAddressClient(cl pbAddress.AddressServiceClient) AddressClient {
	return &addressClient{client: cl}
}

func mapToPBAddress(addr Address) *pbAddress.Address {
	return &pbAddress.Address{
		PublicId: addr.PublicID,
		Location: &pbAddress.Location{
			AddressText: addr.Location.AddressText,
			Latitude:    addr.Location.Latitude,
			Longitude:   addr.Location.Longitude,
		},
		Apartment:      addr.Apartment,
		Entrance:       addr.Entrance,
		Floor:          addr.Floor,
		DoorCode:       addr.DoorCode,
		CourierComment: addr.CourierComment,
		Label:          addr.Label,
	}
}

func mapFromPBAddress(pbAddr *pbAddress.Address) Address {
	if pbAddr == nil {
		return Address{}
	}
	addr := Address{
		PublicID:       pbAddr.PublicId,
		Apartment:      pbAddr.Apartment,
		Entrance:       pbAddr.Entrance,
		Floor:          pbAddr.Floor,
		DoorCode:       pbAddr.DoorCode,
		CourierComment: pbAddr.CourierComment,
		Label:          pbAddr.Label,
	}
	if pbAddr.Location != nil {
		addr.Location = Location{
			AddressText: pbAddr.Location.AddressText,
			Latitude:    pbAddr.Location.Latitude,
			Longitude:   pbAddr.Location.Longitude,
		}
	}
	return addr
}

func (c *addressClient) AddAddress(ctx context.Context, userID int64, addr Address, idempotencyKey string) (string, error) {
	resp, err := c.client.AddAddress(ctx, &pbAddress.AddAddressRequest{
		UserId:         userID,
		Address:        mapToPBAddress(addr),
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return "", ErrInternal
	}
	return resp.AddressPublicId, nil
}

func (c *addressClient) GetMyAddresses(ctx context.Context, userID int64) ([]Address, error) {
	resp, err := c.client.GetMyAddresses(ctx, &pbAddress.GetMyAddressesRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, ErrInternal
	}

	addresses := make([]Address, 0, len(resp.Addresses))
	for _, pbAddr := range resp.Addresses {
		addresses = append(addresses, mapFromPBAddress(pbAddr))
	}
	return addresses, nil
}

func (c *addressClient) DeleteAddress(ctx context.Context, userID int64, publicID string, idempotencyKey string) error {
	_, err := c.client.DeleteAddress(ctx, &pbAddress.DeleteAddressRequest{
		UserId:          userID,
		AddressPublicId: publicID,
		IdempotencyKey:  idempotencyKey,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return ErrAddressNotFound
		}
		return ErrInternal
	}
	return nil
}

func (c *addressClient) UpdateAddress(ctx context.Context, userID int64, addr Address, idempotencyKey string) error {
	_, err := c.client.UpdateAddress(ctx, &pbAddress.UpdateAddressRequest{
		UserId:         userID,
		Address:        mapToPBAddress(addr),
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return ErrAddressNotFound
		}
		return ErrInternal
	}
	return nil
}

func (c *addressClient) CheckAddressExists(ctx context.Context, userID int64, publicID string) error {
	_, err := c.client.CheckAddressExists(ctx, &pbAddress.CheckAddressExistsRequest{
		UserId:          userID,
		AddressPublicId: publicID,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return ErrAddressNotFound
		}
		return ErrInternal
	}
	return nil
}
