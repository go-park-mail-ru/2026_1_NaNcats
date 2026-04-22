package cartclient

import (
	"context"
	"errors"

	pbCart "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/cart"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInvalidCart = errors.New("invalid cart data (wrong quantity or multiple restaurants)")
	ErrInternal    = errors.New("internal server error")
)

type Item struct {
	DishID   int64
	Quantity int32
	Name     string
	Price    int64
	ImageURL string
}

type Cart struct {
	RestaurantID int64
	Items        []Item
	TotalCost    int64
}

type CartClient interface {
	GetCart(ctx context.Context, userID int64) (*Cart, error)
	UpdateCart(ctx context.Context, userID, restaurantID int64, items []Item, idempotencyKey string) error
}

type cartClient struct {
	client pbCart.CartServiceClient
}

func NewCartClient(cl pbCart.CartServiceClient) CartClient {
	return &cartClient{client: cl}
}

func (c *cartClient) GetCart(ctx context.Context, userID int64) (*Cart, error) {
	resp, err := c.client.GetCart(ctx, &pbCart.GetCartRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, ErrInternal
	}

	result := &Cart{
		RestaurantID: resp.Cart.RestaurantBrandId,
		TotalCost:    resp.TotalCost,
		Items:        make([]Item, 0, len(resp.Cart.Items)),
	}

	for _, it := range resp.Cart.Items {
		result.Items = append(result.Items, Item{
			DishID:   it.DishId,
			Quantity: it.Quantity,
			Name:     it.Name,
			Price:    it.Price,
			ImageURL: it.ImageUrl,
		})
	}

	return result, nil
}

func (c *cartClient) UpdateCart(ctx context.Context, userID, restaurantID int64, items []Item, idempotencyKey string) error {
	pbItems := make([]*pbCart.CartItem, 0, len(items))
	for _, it := range items {
		pbItems = append(pbItems, &pbCart.CartItem{
			DishId:   it.DishID,
			Quantity: it.Quantity,
		})
	}

	req := &pbCart.UpdateCartRequest{
		UserId: userID,
		CartData: &pbCart.Cart{
			RestaurantBrandId: restaurantID,
			Items:             pbItems,
		},
		IdempotencyKey: idempotencyKey,
	}

	_, err := c.client.UpdateCart(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.InvalidArgument {
			return ErrInvalidCart
		}
		return ErrInternal
	}

	return nil
}
