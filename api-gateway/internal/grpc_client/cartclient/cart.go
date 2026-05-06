package cartclient

import (
	"context"
	"errors"

	pbCart "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/cart"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:generate mockgen -destination=mocks/cart_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/cartclient CartClient
var (
	ErrInvalidCart         = errors.New("invalid cart data (wrong quantity or multiple restaurants)")
	ErrMultipleRestaurants = errors.New("dish belongs to a different restaurant")
	ErrInvalidQuantity     = errors.New("invalid dish quantity")
	ErrCartLocked          = errors.New("cart is locked")
	ErrForbidden           = errors.New("forbidden")
	ErrNotFound            = errors.New("not found")
	ErrInternal            = errors.New("internal server error")
)

// Внутренние модели BFF

type Item struct {
	DishID      int64
	Quantity    int32
	Name        string
	Price       int64
	ImageURL    string
	OwnerUserID *int64
}

type Member struct {
	UserID   int64
	JoinedAt string
}

type Cart struct {
	ID                string
	AdminID           int64
	RestaurantBrandID int64
	Mode              string
	Status            string
	Items             []Item
	Members           []Member
	TotalCost         int64
}

type InviteResponse struct {
	Token     string
	ExpiresAt string
}

// Интерфейс

type CartClient interface {
	GetCart(ctx context.Context, userID int64) (*Cart, error)
	LockCart(ctx context.Context, cartID string, userID int64, payForAll bool, payerMapping map[int64]int64, idempotencyKey string) error
	UnlockCart(ctx context.Context, cartID string, userID int64, idempotencyKey string) error
	ClearCart(ctx context.Context, cartID string, userID int64, idempotencyKey string) error

	GenerateInvite(ctx context.Context, cartID string, adminID int64) (*InviteResponse, error)
	JoinCart(ctx context.Context, token string, userID int64) (string, error)
	KickMember(ctx context.Context, cartID string, adminID, targetUserID int64, idempotencyKey string) error
	CloseSharedCart(ctx context.Context, cartID string, adminID int64, idempotencyKey string) error

	AddItem(ctx context.Context, cartID string, userID, dishID int64, quantity int32, idempotencyKey string) error
	RemoveItem(ctx context.Context, cartID string, userID, dishID int64, idempotencyKey string) error
	UpdateItemQuantity(ctx context.Context, cartID string, userID, dishID int64, quantity int32, idempotencyKey string) error
	ReassignItemOwner(ctx context.Context, cartID string, adminID, dishID int64, newOwnerID *int64, idempotencyKey string) error
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
		return nil, mapError(err)
	}

	result := &Cart{
		ID:                resp.Cart.CartId,
		AdminID:           resp.Cart.AdminId,
		RestaurantBrandID: resp.Cart.RestaurantBrandId,
		TotalCost:         resp.TotalCost,
		Items:             make([]Item, 0, len(resp.Cart.Items)),
		Members:           make([]Member, 0, len(resp.Cart.Members)),
	}

	if resp.Cart.Mode == pbCart.CartMode_CART_MODE_SHARED {
		result.Mode = "shared"
	} else {
		result.Mode = "solo"
	}

	if resp.Cart.Status == pbCart.CartStatus_CART_STATUS_LOCKED {
		result.Status = "locked"
	} else {
		result.Status = "active"
	}

	for _, it := range resp.Cart.Items {
		result.Items = append(result.Items, Item{
			DishID:      it.DishId,
			Quantity:    it.Quantity,
			Name:        it.Name,
			Price:       it.Price,
			ImageURL:    it.ImageUrl,
			OwnerUserID: it.OwnerUserId,
		})
	}

	for _, m := range resp.Cart.Members {
		result.Members = append(result.Members, Member{
			UserID:   m.UserId,
			JoinedAt: m.JoinedAt,
		})
	}

	return result, nil
}

func (c *cartClient) LockCart(ctx context.Context, cartID string, userID int64, payForAll bool, payerMapping map[int64]int64, idempotencyKey string) error {
	_, err := c.client.LockCart(ctx, &pbCart.LockCartRequest{
		CartId:         cartID,
		UserId:         userID,
		PayForAll:      payForAll,
		PayerMapping:   payerMapping,
		IdempotencyKey: idempotencyKey,
	})
	return mapError(err)
}

func (c *cartClient) UnlockCart(ctx context.Context, cartID string, userID int64, idempotencyKey string) error {
	_, err := c.client.UnlockCart(ctx, &pbCart.CartOperationRequest{CartId: cartID, UserId: userID, IdempotencyKey: idempotencyKey})
	return mapError(err)
}

func (c *cartClient) ClearCart(ctx context.Context, cartID string, userID int64, idempotencyKey string) error {
	_, err := c.client.ClearCart(ctx, &pbCart.CartOperationRequest{CartId: cartID, UserId: userID, IdempotencyKey: idempotencyKey})
	return mapError(err)
}

func (c *cartClient) GenerateInvite(ctx context.Context, cartID string, adminID int64) (*InviteResponse, error) {
	resp, err := c.client.GenerateInvite(ctx, &pbCart.GenerateInviteRequest{CartId: cartID, UserId: adminID})
	if err != nil {
		return nil, mapError(err)
	}
	return &InviteResponse{Token: resp.Token, ExpiresAt: resp.ExpiresAt}, nil
}

func (c *cartClient) JoinCart(ctx context.Context, token string, userID int64) (string, error) {
	resp, err := c.client.JoinCart(ctx, &pbCart.JoinCartRequest{Token: token, UserId: userID})
	if err != nil {
		return "", mapError(err)
	}
	return resp.CartId, nil
}

func (c *cartClient) KickMember(ctx context.Context, cartID string, adminID, targetUserID int64, idempotencyKey string) error {
	_, err := c.client.KickMember(ctx, &pbCart.CartMemberOperationRequest{
		CartId:         cartID,
		AdminUserId:    adminID,
		TargetUserId:   targetUserID,
		IdempotencyKey: idempotencyKey,
	})
	return mapError(err)
}

func (c *cartClient) CloseSharedCart(ctx context.Context, cartID string, adminID int64, idempotencyKey string) error {
	_, err := c.client.CloseSharedCart(ctx, &pbCart.CartOperationRequest{CartId: cartID, UserId: adminID, IdempotencyKey: idempotencyKey})
	return mapError(err)
}

func (c *cartClient) AddItem(ctx context.Context, cartID string, userID, dishID int64, quantity int32, idempotencyKey string) error {
	_, err := c.client.AddItem(ctx, &pbCart.AddItemRequest{
		CartId:         cartID,
		UserId:         userID,
		DishId:         dishID,
		Quantity:       quantity,
		IdempotencyKey: idempotencyKey,
	})
	return mapError(err)
}

func (c *cartClient) RemoveItem(ctx context.Context, cartID string, userID, dishID int64, idempotencyKey string) error {
	_, err := c.client.RemoveItem(ctx, &pbCart.RemoveItemRequest{
		CartId:         cartID,
		UserId:         userID,
		DishId:         dishID,
		IdempotencyKey: idempotencyKey,
	})
	return mapError(err)
}

func (c *cartClient) UpdateItemQuantity(ctx context.Context, cartID string, userID, dishID int64, quantity int32, idempotencyKey string) error {
	_, err := c.client.UpdateItemQuantity(ctx, &pbCart.UpdateQuantityRequest{
		CartId:         cartID,
		UserId:         userID,
		DishId:         dishID,
		NewQuantity:    quantity,
		IdempotencyKey: idempotencyKey,
	})
	return mapError(err)
}

func (c *cartClient) ReassignItemOwner(ctx context.Context, cartID string, adminID, dishID int64, newOwnerID *int64, idempotencyKey string) error {
	_, err := c.client.ReassignItemOwner(ctx, &pbCart.ReassignOwnerRequest{
		CartId:         cartID,
		AdminUserId:    adminID,
		DishId:         dishID,
		NewOwnerUserId: newOwnerID,
		IdempotencyKey: idempotencyKey,
	})
	return mapError(err)
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	st, ok := status.FromError(err)
	if !ok {
		return ErrInternal
	}
	switch st.Code() {
	case codes.InvalidArgument:
		// gRPC status message содержит slug из errutil (см. grpcutil.ToGRPCError):
		// MULTIPLE_RESTAURANTS / INVALID_QUANTITY и т.п.
		switch st.Message() {
		case "MULTIPLE_RESTAURANTS":
			return ErrMultipleRestaurants
		case "INVALID_QUANTITY":
			return ErrInvalidQuantity
		default:
			return ErrInvalidCart
		}
	case codes.FailedPrecondition:
		if st.Message() == "CART_LOCKED" {
			return ErrCartLocked
		}
		return ErrInternal
	case codes.PermissionDenied:
		return ErrForbidden
	case codes.NotFound:
		return ErrNotFound
	default:
		return ErrInternal
	}
}
