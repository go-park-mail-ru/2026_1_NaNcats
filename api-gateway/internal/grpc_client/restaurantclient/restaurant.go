package restaurantclient

import (
	"context"
	"errors"
	"net/url"
	"strconv"

	pbRestaurant "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/restaurant"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	ErrNotFound = errors.New("restaurant or dish not found")
	ErrInternal = errors.New("internal server error")
)

//go:generate mockgen -destination=mocks/restaurant_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/restaurantclient RestaurantClient
type RestaurantClient interface {
	GetRestaurantBrandsList(ctx context.Context, limit, offset int32) ([]*pbRestaurant.RestaurantBrand, error)
	GetRestaurantBrandByID(ctx context.Context, id int64) (*pbRestaurant.RestaurantBrand, error)
	GetDishesByRestaurantBrandID(ctx context.Context, brandID int64, limit, offset int32) ([]*pbRestaurant.Dish, error)
	GetDishesByIDs(ctx context.Context, dishIDs []int64) ([]*pbRestaurant.Dish, error)
	GetRestaurantLogos(ctx context.Context, brandIDs []int64) (map[int64]string, error)
	GetRestaurantBrandsListByCategory(ctx context.Context, categoryID int64, limit, offset int32) ([]*pbRestaurant.RestaurantBrand, error)
	GetRestaurantBrandsListByCategoryName(ctx context.Context, categoryName string, limit, offset int32) ([]*pbRestaurant.RestaurantBrand, error)
	SearchRestaurantBrands(ctx context.Context, query string, limit, offset int32) ([]*pbRestaurant.RestaurantBrand, error)
	SearchDishes(ctx context.Context, query string, limit int32) ([]*pbRestaurant.Dish, error)
	SearchDishesByBrand(ctx context.Context, brandID int64, query string, limit int32) ([]*pbRestaurant.Dish, error)
	GetDishByID(ctx context.Context, dishID int64) (*pbRestaurant.Dish, error)
	DeleteRestaurantBrand(ctx context.Context, id int64) error
	UpdateRestaurantBrand(ctx context.Context, id int64, name, desc *string, logo []byte, tier *int32, idemKey string) (*pbRestaurant.RestaurantBrand, error)
	CreateRestaurantBrand(ctx context.Context, ownerID int64, name, desc string, logo []byte, idemKey string) (*pbRestaurant.RestaurantBrand, error)
	DeleteDish(ctx context.Context, id int64) error
	UpdateDish(ctx context.Context, id int64, name, desc *string, price *int64, image []byte, idemKey string) (*pbRestaurant.Dish, error)
	CreateDish(ctx context.Context, brandID int64, name, desc string, price int64, image []byte, idemKey string) (*pbRestaurant.Dish, error)
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

func (c *restaurantClient) GetRestaurantBrandsListByCategory(ctx context.Context, categoryID int64, limit, offset int32) ([]*pbRestaurant.RestaurantBrand, error) {
	md := metadata.New(map[string]string{"x-category-id": strconv.FormatInt(categoryID, 10)})
	ctx = metadata.NewOutgoingContext(ctx, md)
	resp, err := c.client.GetRestaurantBrandsList(ctx, &pbRestaurant.GetRestaurantBrandsListRequest{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, ErrInternal
	}
	return resp.RestaurantBrands, nil
}

func (c *restaurantClient) GetRestaurantBrandsListByCategoryName(ctx context.Context, categoryName string, limit, offset int32) ([]*pbRestaurant.RestaurantBrand, error) {
	md := metadata.New(map[string]string{"x-category-name": url.QueryEscape(categoryName)})
	ctx = metadata.NewOutgoingContext(ctx, md)
	resp, err := c.client.GetRestaurantBrandsList(ctx, &pbRestaurant.GetRestaurantBrandsListRequest{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, ErrInternal
	}
	return resp.RestaurantBrands, nil
}

func (c *restaurantClient) SearchRestaurantBrands(ctx context.Context, query string, limit, offset int32) ([]*pbRestaurant.RestaurantBrand, error) {
	md := metadata.New(map[string]string{"x-search-query": url.QueryEscape(query)})
	ctx = metadata.NewOutgoingContext(ctx, md)
	resp, err := c.client.GetRestaurantBrandsList(ctx, &pbRestaurant.GetRestaurantBrandsListRequest{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, ErrInternal
	}
	return resp.RestaurantBrands, nil
}

func (c *restaurantClient) SearchDishes(ctx context.Context, query string, limit int32) ([]*pbRestaurant.Dish, error) {
	md := metadata.New(map[string]string{
		"x-dish-search":       url.QueryEscape(query),
		"x-dish-search-limit": strconv.Itoa(int(limit)),
	})
	ctx = metadata.NewOutgoingContext(ctx, md)
	resp, err := c.client.GetDishesByIDs(ctx, &pbRestaurant.GetDishesByIDsRequest{DishIds: []int64{}})
	if err != nil {
		return nil, ErrInternal
	}
	return resp.Dishes, nil
}

func (c *restaurantClient) SearchDishesByBrand(ctx context.Context, brandID int64, query string, limit int32) ([]*pbRestaurant.Dish, error) {
	md := metadata.New(map[string]string{"x-dish-search": url.QueryEscape(query)})
	ctx = metadata.NewOutgoingContext(ctx, md)
	resp, err := c.client.GetDishesByRestaurantBrandID(ctx, &pbRestaurant.GetDishesByRestaurantBrandIDRequest{
		RestaurantBrandId: brandID,
		Limit:             limit,
		Offset:            0,
	})
	if err != nil {
		return nil, ErrInternal
	}
	return resp.Dishes, nil
}
