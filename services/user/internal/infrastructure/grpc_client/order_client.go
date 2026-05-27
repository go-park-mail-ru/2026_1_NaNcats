package grpc_client

import (
	"context"
	"fmt"

	pbOrder "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/order"
)

type orderClient struct {
	orderClient pbOrder.OrderServiceClient
	promoClient pbOrder.PromoServiceClient
}

func NewOrderClient(oc pbOrder.OrderServiceClient, pc pbOrder.PromoServiceClient) *orderClient {
	return &orderClient{
		orderClient: oc,
		promoClient: pc,
	}
}

func (c *orderClient) GetTrendingBrands(ctx context.Context, windowDays, limit int32) ([]int64, error) {
	resp, err := c.orderClient.GetTrendingBrands(ctx, &pbOrder.GetTrendingBrandsRequest{
		WindowDays: windowDays,
		Limit:      limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch trending brands over gRPC: %w", err)
	}
	return resp.BrandIds, nil
}

func (c *orderClient) CreateAndBindWheelPromo(
	ctx context.Context,
	userID int64,
	title string,
	discountAmount *int64,
	discountPercent *int,
	brandID *int64,
	minOrderAmount *int64,
) (string, *string, error) {

	req := &pbOrder.CreateAndBindWheelPromoRequest{
		UserId: userID,
		Title:  title,
	}

	if discountAmount != nil {
		req.DiscountAmount = discountAmount
	}
	if discountPercent != nil {
		val := int32(*discountPercent)
		req.DiscountPercent = &val
	}
	if brandID != nil {
		req.RestaurantBrandId = brandID
	}
	if minOrderAmount != nil {
		req.MinOrderAmount = minOrderAmount
	}

	resp, err := c.promoClient.CreateAndBindWheelPromo(ctx, req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create wheel promo over gRPC: %w", err)
	}

	return resp.Code, resp.ExpiresAt, nil
}
