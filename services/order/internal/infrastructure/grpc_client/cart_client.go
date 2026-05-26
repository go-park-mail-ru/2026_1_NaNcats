package grpc_client

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/usecase"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/cart"
)

type cartClient struct {
	client pb.CartServiceClient
}

func NewCartClient(cl pb.CartServiceClient) usecase.CartClient {
	return &cartClient{
		client: cl,
	}
}

func (c *cartClient) GetCart(ctx context.Context, userID int64) (domain.Cart, int64, error) {
	resp, err := c.client.GetCart(ctx, &pb.GetCartRequest{
		UserId: userID,
	})
	if err != nil {
		return domain.Cart{}, 0, err
	}

	items := make([]domain.CartItem, 0, len(resp.Cart.Items))
	for _, item := range resp.Cart.Items {
		items = append(items, domain.CartItem{
			DishID:      item.DishId,
			Name:        item.Name,
			Quantity:    int(item.Quantity),
			Price:       item.Price,
			OwnerUserID: item.OwnerUserId,
		})
	}

	cart := domain.Cart{
		ID:                resp.Cart.CartId,
		RestaurantBrandID: resp.Cart.RestaurantBrandId,
		Items:             items,
	}

	return cart, resp.TotalCost, nil
}

func (c *cartClient) LockCart(ctx context.Context, cartID string, userID int64, idempotencyKey string) error {
	_, err := c.client.LockCart(ctx, &pb.LockCartRequest{
		CartId:         cartID,
		UserId:         userID,
		IdempotencyKey: idempotencyKey,
	})
	return err
}

func (c *cartClient) UnlockCart(ctx context.Context, cartID string, userID int64, idempotencyKey string) error {
	_, err := c.client.UnlockCart(ctx, &pb.CartOperationRequest{
		CartId:         cartID,
		UserId:         userID,
		IdempotencyKey: idempotencyKey,
	})
	return err
}
