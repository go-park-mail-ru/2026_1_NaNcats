package grpc_client

import (
	"context"
	"fmt"

	analyticsDomain "github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/domain"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/restaurant"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type restaurantClient struct {
	client pb.RestaurantServiceClient
}

func NewRestaurantClient(cl pb.RestaurantServiceClient) *restaurantClient {
	return &restaurantClient{
		client: cl,
	}
}

func (c *restaurantClient) GetRestaurantBrandByID(ctx context.Context, id int64) (analyticsDomain.RestaurantBrand, error) {
	resp, err := c.client.GetRestaurantBrandByID(ctx, &pb.GetRestaurantBrandByIDRequest{
		Id: id,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return analyticsDomain.RestaurantBrand{}, analyticsDomain.ErrRestaurantNotFound
		}
		return analyticsDomain.RestaurantBrand{}, fmt.Errorf("failed to fetch restaurant brand over grpc: %w", err)
	}

	if resp == nil || resp.RestaurantBrand == nil {
		return analyticsDomain.RestaurantBrand{}, analyticsDomain.ErrRestaurantNotFound
	}

	return analyticsDomain.RestaurantBrand{
		ID:             resp.RestaurantBrand.Id,
		OwnerProfileID: resp.RestaurantBrand.OwnerProfileId,
		Name:           resp.RestaurantBrand.Name,
	}, nil
}
