package grpc_client

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/usecase"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/order"
)

type orderClient struct {
	client pb.OrderServiceClient
}

func NewOrderClient(cl pb.OrderServiceClient) usecase.OrderHistoryClient {
	return &orderClient{client: cl}
}

func (c *orderClient) GetUserPaidBrands(ctx context.Context, userID int64) ([]int64, error) {
	resp, err := c.client.GetUserPaidBrands(ctx, &pb.GetUserPaidBrandsRequest{UserId: userID})
	if err != nil {
		return nil, err
	}
	return resp.BrandIds, nil
}

func (c *orderClient) GetTrendingBrands(ctx context.Context, windowDays, limit int32) ([]int64, error) {
	resp, err := c.client.GetTrendingBrands(ctx, &pb.GetTrendingBrandsRequest{
		WindowDays: windowDays,
		Limit:      limit,
	})
	if err != nil {
		return nil, err
	}
	return resp.BrandIds, nil
}

func (c *orderClient) GetTopDishesByBrand(ctx context.Context, brandID int64, windowDays, limit int32) ([]int64, error) {
	resp, err := c.client.GetTopDishesByBrand(ctx, &pb.GetTopDishesByBrandRequest{
		BrandId:    brandID,
		WindowDays: windowDays,
		Limit:      limit,
	})
	if err != nil {
		return nil, err
	}
	return resp.DishIds, nil
}
