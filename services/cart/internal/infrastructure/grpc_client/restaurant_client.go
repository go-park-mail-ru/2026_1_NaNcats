package grpc_client

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/usecase"
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

func (c *restaurantClient) GetDishesByIDs(ctx context.Context, dishIDs []int64) ([]domain.Dish, error) {
	resp, err := c.client.GetDishesByIDs(ctx, &pb.GetDishesByIDsRequest{
		DishIds: dishIDs,
	})
	if err != nil {
		return nil, err
	}

	dishes := make([]domain.Dish, 0, len(resp.Dishes))
	for _, d := range resp.Dishes {
		dishes = append(dishes, domain.Dish{
			ID:                d.Id,
			Name:              d.Name,
			Price:             d.Price,
			ImageURL:          d.ImageUrl,
			RestaurantBrandID: d.RestaurantBrandId,
		})
	}

	return dishes, nil
}
