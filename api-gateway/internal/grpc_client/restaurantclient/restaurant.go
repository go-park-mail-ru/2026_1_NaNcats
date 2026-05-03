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
	resp, err := c.client.GetRestaurantBrandsByIDs(ctx, &pbRestaurant.GetRestaurantBrandsByIDsRequest{
		BrandIds: brandIDs,
	})
	if err != nil {
		return nil, ErrInternal
	}
	logos := make(map[int64]string, len(resp.RestaurantBrands))
	for _, restaurantBrand := range resp.RestaurantBrands {
		logos[restaurantBrand.Id] = restaurantBrand.LogoUrl
	}

	return logos, nil
}

func (c *restaurantClient) GetDishByID(ctx context.Context, dishID int64) (*pbRestaurant.Dish, error) {
	resp, err := c.client.GetDishesByIDs(ctx, &pbRestaurant.GetDishesByIDsRequest{
		DishIds: []int64{dishID},
	})
	if err != nil || len(resp.Dishes) == 0 {
		return nil, ErrNotFound
	}
	return resp.Dishes[0], nil
}

func (c *restaurantClient) CreateRestaurantBrand(ctx context.Context, ownerID int64, name, desc string, logo []byte, idemKey string) (*pbRestaurant.RestaurantBrand, error) {
	resp, err := c.client.CreateRestaurantBrand(ctx, &pbRestaurant.CreateBrandRequest{
		OwnerId:        ownerID,
		Name:           name,
		Description:    desc,
		LogoData:       logo,
		IdempotencyKey: idemKey,
	})
	return resp, err
}

func (c *restaurantClient) UpdateRestaurantBrand(ctx context.Context, id int64, name, desc *string, logo []byte, tier *int32, idemKey string) (*pbRestaurant.RestaurantBrand, error) {
	req := &pbRestaurant.UpdateBrandRequest{
		Id:             id,
		LogoData:       logo,
		IdempotencyKey: idemKey,
	}
	if name != nil {
		req.Name = name
	}
	if desc != nil {
		req.Description = desc
	}
	if tier != nil {
		req.PromotionTier = tier
	}

	return c.client.UpdateRestaurantBrand(ctx, req)
}

func (c *restaurantClient) DeleteRestaurantBrand(ctx context.Context, id int64) error {
	_, err := c.client.DeleteRestaurantBrand(ctx, &pbRestaurant.DeleteBrandRequest{Id: id})
	return err
}

func (c *restaurantClient) CreateDish(ctx context.Context, brandID int64, name, desc string, price int64, image []byte, idemKey string) (*pbRestaurant.Dish, error) {
	resp, err := c.client.CreateDish(ctx, &pbRestaurant.CreateDishRequest{
		RestaurantBrandId: brandID,
		Name:              name,
		Description:       desc,
		Price:             price,
		ImageData:         image,
		IdempotencyKey:    idemKey,
	})
	return resp, err
}

func (c *restaurantClient) UpdateDish(ctx context.Context, id int64, name, desc *string, price *int64, image []byte, idemKey string) (*pbRestaurant.Dish, error) {
	req := &pbRestaurant.UpdateDishRequest{
		Id:             id,
		ImageData:      image,
		IdempotencyKey: idemKey,
	}
	if name != nil {
		req.Name = name
	}
	if desc != nil {
		req.Description = desc
	}
	if price != nil {
		req.Price = price
	}

	return c.client.UpdateDish(ctx, req)
}

func (c *restaurantClient) DeleteDish(ctx context.Context, id int64) error {
	_, err := c.client.DeleteDish(ctx, &pbRestaurant.DeleteDishRequest{Id: id})
	return err
}
