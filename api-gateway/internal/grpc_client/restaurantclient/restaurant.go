package restaurantclient

import (
	"context"
	"errors"

	pbRestaurant "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/restaurant"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrNotFound = errors.New("restaurant or dish not found")
	ErrInternal = errors.New("internal server error")
)

type RestaurantClient interface {
	GetRestaurantBrandsList(ctx context.Context, limit, offset int32) ([]*pbRestaurant.RestaurantBrand, error)
	GetRestaurantBrandByID(ctx context.Context, id int64) (*pbRestaurant.RestaurantBrand, error)
	GetDishesByRestaurantBrandID(ctx context.Context, brandID int64, limit, offset int32) ([]*pbRestaurant.Dish, error)
	GetDishesByIDs(ctx context.Context, dishIDs []int64) ([]*pbRestaurant.Dish, error)
	GetRestaurantLogos(ctx context.Context, brandIDs []int64) (map[int64]string, error)
}

type restaurantClient struct {
	client pbRestaurant.RestaurantServiceClient
}

func NewRestaurantClient(cl pbRestaurant.RestaurantServiceClient) RestaurantClient {
	return &restaurantClient{client: cl}
}

func (c *restaurantClient) GetRestaurantBrandsList(ctx context.Context, limit, offset int32) ([]*pbRestaurant.RestaurantBrand, error) {
	resp, err := c.client.GetRestaurantBrandsList(ctx, &pbRestaurant.GetRestaurantBrandsListRequest{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, ErrInternal
	}
	return resp.RestaurantBrands, nil
}

func (c *restaurantClient) GetRestaurantBrandByID(ctx context.Context, id int64) (*pbRestaurant.RestaurantBrand, error) {
	resp, err := c.client.GetRestaurantBrandByID(ctx, &pbRestaurant.GetRestaurantBrandByIDRequest{
		Id: id,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}
	return resp.RestaurantBrand, nil
}

func (c *restaurantClient) GetDishesByRestaurantBrandID(ctx context.Context, brandID int64, limit, offset int32) ([]*pbRestaurant.Dish, error) {
	resp, err := c.client.GetDishesByRestaurantBrandID(ctx, &pbRestaurant.GetDishesByRestaurantBrandIDRequest{
		RestaurantBrandId: brandID,
		Limit:             limit,
		Offset:            offset,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}
	return resp.Dishes, nil
}

func (c *restaurantClient) GetDishesByIDs(ctx context.Context, dishIDs []int64) ([]*pbRestaurant.Dish, error) {
	resp, err := c.client.GetDishesByIDs(ctx, &pbRestaurant.GetDishesByIDsRequest{
		DishIds: dishIDs,
	})
	if err != nil {
		return nil, ErrInternal
	}
	return resp.Dishes, nil
}

func (c *restaurantClient) GetRestaurantLogos(ctx context.Context, brandIDs []int64) (map[int64]string, error) {
	resp, err := c.client.GetRestaurantLogos(ctx, &pbRestaurant.GetRestaurantLogosRequest{
		BrandIds: brandIDs,
	})
	if err != nil {
		return nil, ErrInternal
	}
	return resp.Logos, nil
}
