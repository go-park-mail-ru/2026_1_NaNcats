package grpc_client

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/usecase"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/restaurant"
)

type restaurantClient struct {
	client pb.RestaurantServiceClient
}

func NewRestaurantClient(cl pb.RestaurantServiceClient) usecase.RestaurantClient {
	return &restaurantClient{
		client: cl,
	}
}

func (c *restaurantClient) GetRestaurantName(ctx context.Context, branchID int64) (string, error) {
	resp, err := c.client.GetRestaurantBrandByID(ctx, &pb.GetRestaurantBrandByIDRequest{
		Id: branchID,
	})
	if err != nil {
		return "", err
	}
	return resp.RestaurantBrand.Name, nil
}

func (c *restaurantClient) GetLogosByBranchIDs(ctx context.Context, branchIDs []int64) (map[int64]string, error) {
	resp, err := c.client.GetRestaurantLogos(ctx, &pb.GetRestaurantLogosRequest{
		BrandIds: branchIDs,
	})
	if err != nil {
		return nil, err
	}

	return resp.Logos, nil
}
