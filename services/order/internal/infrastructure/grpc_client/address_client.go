package grpc_client

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/usecase"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/address"
)

type addressClient struct {
	client pb.AddressServiceClient
}

func NewAddressClient(cl pb.AddressServiceClient) usecase.AddressClient {
	return &addressClient{
		client: cl,
	}
}

func (c *addressClient) CheckAddressExists(ctx context.Context, userID int64, addressPublicID string) error {
	_, err := c.client.CheckAddressExists(ctx, &pb.CheckAddressExistsRequest{
		UserId:          userID,
		AddressPublicId: addressPublicID,
	})

	return err
}
