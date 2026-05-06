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

func (c *restaurantClient) GetLogosByBrandIDs(ctx context.Context, brandIDs []int64) (map[int64]string, error) {
	resp, err := c.client.GetRestaurantBrandsByIDs(ctx, &pb.GetRestaurantBrandsByIDsRequest{
		BrandIds: brandIDs,
	})
	if err != nil {
		return nil, err
	}
	logos := make(map[int64]string, len(resp.RestaurantBrands))
	for _, restaurantBrand := range resp.RestaurantBrands {
		logos[restaurantBrand.Id] = restaurantBrand.LogoUrl
	}

	return logos, nil
}
