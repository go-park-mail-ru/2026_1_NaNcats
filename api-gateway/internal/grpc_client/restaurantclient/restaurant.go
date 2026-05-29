package restaurantclient

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"

	pbRestaurant "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/restaurant"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func mapPBRestaurant(pb *pbRestaurant.RestaurantBrand) RestaurantBrand {
	if pb == nil {
		return RestaurantBrand{}
	}
	return RestaurantBrand{
		ID:             pb.Id,
		OwnerProfileID: pb.OwnerProfileId,
		Name:           pb.Name,
		Description:    pb.Description,
		PromotionTier:  int(pb.PromotionTier),
		LogoURL:        pb.LogoUrl,
	}
}

func mapPBDish(pb *pbRestaurant.Dish) Dish {
	if pb == nil {
		return Dish{}
	}
	return Dish{
		ID:                pb.Id,
		RestaurantBrandID: pb.RestaurantBrandId,
		Name:              pb.Name,
		Description:       pb.Description,
		ImageURL:          pb.ImageUrl,
		Price:             pb.Price,
		Section:           pb.Section,
	}
}

func mapPBCategory(pb *pbRestaurant.Category) Category {
	if pb == nil {
		return Category{}
	}
	return Category{
		ID:    pb.Id,
		Name:  pb.Name,
		Emoji: pb.Emoji,
	}
}

var (
	ErrNotFound = errors.New("restaurant or dish not found")
	ErrInternal = errors.New("internal server error")
)

type Category struct {
	ID    int64
	Name  string
	Emoji string
}

type RestaurantBrand struct {
	ID             int64
	OwnerProfileID int64
	Name           string
	Description    string
	PromotionTier  int
	LogoURL        string
}

type Dish struct {
	ID                int64
	RestaurantBrandID int64
	Name              string
	Description       string
	ImageURL          string
	Price             int64
	Section           string
}

//go:generate mockgen -destination=mocks/restaurant_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/restaurantclient RestaurantClient
type RestaurantClient interface {
	GetCategories(ctx context.Context) ([]Category, error)
	GetRestaurantBrandsList(ctx context.Context, limit, offset int32) ([]RestaurantBrand, error)
	GetRestaurantBrandByID(ctx context.Context, id int64) (RestaurantBrand, error)
	GetDishesByRestaurantBrandID(ctx context.Context, brandID int64, limit, offset int32) ([]Dish, error)
	GetDishesByIDs(ctx context.Context, dishIDs []int64) ([]Dish, error)
	GetRestaurantLogos(ctx context.Context, brandIDs []int64) (map[int64]string, error)
	GetRestaurantBrandsByCategoryName(ctx context.Context, categoryName string, limit, offset int32) ([]RestaurantBrand, error)
	SearchRestaurantBrands(ctx context.Context, query string, limit, offset int32) ([]RestaurantBrand, error)
	SearchDishes(ctx context.Context, query string, limit int32) ([]Dish, error)
	SearchDishesByBrand(ctx context.Context, brandID int64, query string, limit int32) ([]Dish, error)
	GetDishByID(ctx context.Context, dishID int64) (Dish, error)
	DeleteRestaurantBrand(ctx context.Context, id int64) error
	UpdateRestaurantBrand(ctx context.Context, id int64, name, desc *string, logo []byte, tier *int32, idemKey string) (RestaurantBrand, error)
	CreateRestaurantBrand(ctx context.Context, ownerID int64, name, desc string, logo []byte, idemKey string) (RestaurantBrand, error)
	DeleteDish(ctx context.Context, id int64) error
	UpdateDish(ctx context.Context, id int64, name, desc *string, price *int64, image []byte, idemKey string) (Dish, error)
	CreateDish(ctx context.Context, brandID int64, name, desc string, price int64, image []byte, idemKey string) (Dish, error)
	GetRecommendations(ctx context.Context, userID int64, limit int32) ([]RestaurantBrand, error)
	GetRecommendedDishes(ctx context.Context, brandID, userID int64, limit int32) ([]Dish, error)
}

type restaurantClient struct {
	client pbRestaurant.RestaurantServiceClient
}

func NewRestaurantClient(cl pbRestaurant.RestaurantServiceClient) RestaurantClient {
	return &restaurantClient{client: cl}
}

func (c *restaurantClient) GetCategories(ctx context.Context) ([]Category, error) {
	resp, err := c.client.GetCategories(ctx, &emptypb.Empty{})
	if err != nil {
		return []Category{}, ErrInternal
	}

	categories := make([]Category, 0, len(resp.Categories))
	for _, cat := range resp.Categories {
		categories = append(categories, mapPBCategory(cat))
	}
	return categories, nil
}

func (c *restaurantClient) GetRestaurantBrandsList(ctx context.Context, limit, offset int32) ([]RestaurantBrand, error) {
	resp, err := c.client.GetRestaurantBrandsList(ctx, &pbRestaurant.GetRestaurantBrandsListRequest{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, ErrInternal
	}

	brands := make([]RestaurantBrand, 0, len(resp.RestaurantBrands))
	for _, b := range resp.RestaurantBrands {
		brands = append(brands, mapPBRestaurant(b))
	}
	return brands, nil
}

func (c *restaurantClient) GetRecommendations(ctx context.Context, userID int64, limit int32) ([]RestaurantBrand, error) {
	resp, err := c.client.GetRecommendations(ctx, &pbRestaurant.GetRecommendationsRequest{
		UserId: userID,
		Limit:  limit,
	})
	if err != nil {
		return nil, ErrInternal
	}
	brands := make([]RestaurantBrand, 0, len(resp.RestaurantBrands))
	for _, b := range resp.RestaurantBrands {
		brands = append(brands, mapPBRestaurant(b))
	}
	return brands, nil
}

func (c *restaurantClient) GetRecommendedDishes(ctx context.Context, brandID, userID int64, limit int32) ([]Dish, error) {
	resp, err := c.client.GetRecommendedDishes(ctx, &pbRestaurant.GetRecommendedDishesRequest{
		BrandId: brandID,
		UserId:  userID,
		Limit:   limit,
	})
	if err != nil {
		return nil, ErrInternal
	}
	dishes := make([]Dish, 0, len(resp.Dishes))
	for _, d := range resp.Dishes {
		dishes = append(dishes, mapPBDish(d))
	}
	return dishes, nil
}

func (c *restaurantClient) GetRestaurantBrandByID(ctx context.Context, id int64) (RestaurantBrand, error) {
	resp, err := c.client.GetRestaurantBrandByID(ctx, &pbRestaurant.GetRestaurantBrandByIDRequest{
		Id: id,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return RestaurantBrand{}, ErrNotFound
		}
		return RestaurantBrand{}, ErrInternal
	}
	return mapPBRestaurant(resp.RestaurantBrand), nil
}

func (c *restaurantClient) GetDishesByRestaurantBrandID(ctx context.Context, brandID int64, limit, offset int32) ([]Dish, error) {
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

	dishes := make([]Dish, 0, len(resp.Dishes))
	for _, d := range resp.Dishes {
		dishes = append(dishes, mapPBDish(d))
	}
	return dishes, nil
}

func (c *restaurantClient) GetDishesByIDs(ctx context.Context, dishIDs []int64) ([]Dish, error) {
	resp, err := c.client.GetDishesByIDs(ctx, &pbRestaurant.GetDishesByIDsRequest{
		DishIds: dishIDs,
	})
	if err != nil {
		return nil, ErrInternal
	}

	dishes := make([]Dish, 0, len(resp.Dishes))
	for _, d := range resp.Dishes {
		dishes = append(dishes, mapPBDish(d))
	}
	return dishes, nil
}

func (c *restaurantClient) GetRestaurantLogos(ctx context.Context, brandIDs []int64) (map[int64]string, error) {
	resp, err := c.client.GetRestaurantBrandsByIDs(ctx, &pbRestaurant.GetRestaurantBrandsByIDsRequest{
		BrandIds: brandIDs,
	})
	if err != nil {
		return nil, ErrInternal
	}

	logos := make(map[int64]string, len(resp.RestaurantBrands))
	for _, b := range resp.RestaurantBrands {
		logos[b.Id] = b.LogoUrl
	}
	return logos, nil
}

func (c *restaurantClient) GetDishByID(ctx context.Context, dishID int64) (Dish, error) {
	resp, err := c.client.GetDishesByIDs(ctx, &pbRestaurant.GetDishesByIDsRequest{
		DishIds: []int64{dishID},
	})
	if err != nil || len(resp.Dishes) == 0 {
		return Dish{}, ErrNotFound
	}
	return mapPBDish(resp.Dishes[0]), nil
}

func (c *restaurantClient) GetRestaurantBrandsByCategoryName(ctx context.Context, categoryName string, limit, offset int32) ([]RestaurantBrand, error) {
	resp, err := c.client.GetRestaurantBrandsByCategory(ctx, &pbRestaurant.GetRestaurantBrandsByCategoryRequest{
		CategoryName: categoryName,
		Limit:        limit,
		Offset:       offset,
	})
	if err != nil {
		return nil, fmt.Errorf("gRPC GetRestaurantBrandsByCategory failed: %w", err)
	}

	brands := make([]RestaurantBrand, 0, len(resp.RestaurantBrands))
	for _, b := range resp.RestaurantBrands {
		brands = append(brands, mapPBRestaurant(b))
	}
	return brands, nil
}

func (c *restaurantClient) CreateRestaurantBrand(ctx context.Context, ownerID int64, name, desc string, logo []byte, idemKey string) (RestaurantBrand, error) {
	pbResp, err := c.client.CreateRestaurantBrand(ctx, &pbRestaurant.CreateBrandRequest{
		OwnerId:        ownerID,
		Name:           name,
		Description:    desc,
		LogoData:       logo,
		IdempotencyKey: idemKey,
	})

	resp := mapPBRestaurant(pbResp)
	return resp, err
}

func (c *restaurantClient) UpdateRestaurantBrand(ctx context.Context, id int64, name, desc *string, logo []byte, tier *int32, idemKey string) (RestaurantBrand, error) {
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

	pbResp, err := c.client.UpdateRestaurantBrand(ctx, req)
	resp := mapPBRestaurant(pbResp)
	return resp, err
}

func (c *restaurantClient) DeleteRestaurantBrand(ctx context.Context, id int64) error {
	_, err := c.client.DeleteRestaurantBrand(ctx, &pbRestaurant.DeleteBrandRequest{Id: id})
	return err
}

func (c *restaurantClient) CreateDish(ctx context.Context, brandID int64, name, desc string, price int64, image []byte, idemKey string) (Dish, error) {
	pbResp, err := c.client.CreateDish(ctx, &pbRestaurant.CreateDishRequest{
		RestaurantBrandId: brandID,
		Name:              name,
		Description:       desc,
		Price:             price,
		ImageData:         image,
		IdempotencyKey:    idemKey,
	})

	resp := mapPBDish(pbResp)
	return resp, err
}

func (c *restaurantClient) UpdateDish(ctx context.Context, id int64, name, desc *string, price *int64, image []byte, idemKey string) (Dish, error) {
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

	pbResp, err := c.client.UpdateDish(ctx, req)
	resp := mapPBDish(pbResp)
	return resp, err
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

func (c *restaurantClient) SearchRestaurantBrands(ctx context.Context, query string, limit, offset int32) ([]RestaurantBrand, error) {
	resp, err := c.client.SearchRestaurantBrands(ctx, &pbRestaurant.SearchRestaurantBrandsRequest{
		Query:  query,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, ErrInternal
	}

	brands := make([]RestaurantBrand, 0, len(resp.RestaurantBrands))
	for _, b := range resp.RestaurantBrands {
		brands = append(brands, mapPBRestaurant(b))
	}
	return brands, nil
}

func (c *restaurantClient) SearchDishes(ctx context.Context, query string, limit int32) ([]Dish, error) {
	resp, err := c.client.SearchDishes(ctx, &pbRestaurant.SearchDishesRequest{
		Query: query,
		Limit: limit,
	})
	if err != nil {
		return nil, ErrInternal
	}

	dishes := make([]Dish, 0, len(resp.Dishes))
	for _, d := range resp.Dishes {
		dishes = append(dishes, mapPBDish(d))
	}
	return dishes, nil
}

func (c *restaurantClient) SearchDishesByBrand(ctx context.Context, brandID int64, query string, limit int32) ([]Dish, error) {
	resp, err := c.client.SearchDishesByBrand(ctx, &pbRestaurant.SearchDishesByBrandRequest{
		RestaurantBrandId: brandID,
		Query:             query,
		Limit:             limit,
	})
	if err != nil {
		return nil, ErrInternal
	}

	dishes := make([]Dish, 0, len(resp.Dishes))
	for _, d := range resp.Dishes {
		dishes = append(dishes, mapPBDish(d))
	}
	return dishes, nil
}
