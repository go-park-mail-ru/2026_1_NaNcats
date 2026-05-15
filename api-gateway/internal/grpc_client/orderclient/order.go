package orderclient

import (
	"context"
	"errors"
	"fmt"
	"time"

	pbOrder "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/order"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:generate mockgen -destination=mocks/order_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/orderclient OrderClient
var (
	ErrCartIsEmpty     = errors.New("cart is empty")
	ErrAddressNotFound = errors.New("address not found")
	ErrUnassignedItems = errors.New("cart has unassigned items")
	ErrInternal        = errors.New("internal server error")
)

type CreateOrderInput struct {
	AddressPublicID    string
	RestaurantBranchID int64
	RestaurantBrandID  int64
	DeliveryCost       int64
	ServiceFee         int64

	PaymentMethodID string
	PayForAll       bool
	PayerMapping    map[int64]int64
	Promocode       *string
}

type OrderDish struct {
	DishID      int64
	DishName    string
	Quantity    int32
	Price       int64
	OwnerUserID *int64
}

type OrderSplit struct {
	SplitID        string
	UserID         int64
	BaseAmount     int64
	DiscountAmount int64
	Amount         int64
	Status         string
}

type Order struct {
	PublicID          string
	RestaurantName    string
	RestaurantBrandID int64
	TotalCost         int64
	Promocode         *string
	DiscountAmount    int64
	Status            string
	CreatedAt         time.Time
	Items             []OrderDish
	Splits            []OrderSplit
}

type OrderClient interface {
	CreateOrder(ctx context.Context, userID int64, input CreateOrderInput, idempotencyKey string) (string, error)
	GetOrders(ctx context.Context, userID int64, limit, offset int32) ([]Order, error)
	PayForFriend(ctx context.Context, splitID string, payerID int64, paymentMethodID, idempotencyKey string) error
	CancelOrder(ctx context.Context, orderPublicID string, userID int64) error
}

type orderClient struct {
	client pbOrder.OrderServiceClient
}

func NewOrderClient(cl pbOrder.OrderServiceClient) OrderClient {
	return &orderClient{client: cl}
}

func (c *orderClient) CreateOrder(ctx context.Context, userID int64, input CreateOrderInput, idempotencyKey string) (string, error) {
	req := &pbOrder.CreateOrderRequest{
		UserId:             userID,
		AddressPublicId:    input.AddressPublicID,
		RestaurantBranchId: input.RestaurantBranchID,
		RestaurantBrandId:  input.RestaurantBrandID,
		PaymentMethodId:    input.PaymentMethodID,
		PayForAll:          input.PayForAll,
		PayerMapping:       input.PayerMapping,
		DeliveryCost:       input.DeliveryCost,
		ServiceFee:         input.ServiceFee,
		IdempotencyKey:     idempotencyKey,
		Promocode:          input.Promocode,
	}

	resp, err := c.client.CreateOrder(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				return "", ErrAddressNotFound
			case codes.FailedPrecondition:
				return "", ErrUnassignedItems
			case codes.InvalidArgument:
				return "", ErrCartIsEmpty
			}
			return "", fmt.Errorf("order service error: %s", st.Message())
		}
		return "", ErrInternal
	}

	return resp.OrderPublicId, nil
}

func (c *orderClient) GetOrders(ctx context.Context, userID int64, limit, offset int32) ([]Order, error) {
	resp, err := c.client.GetOrders(ctx, &pbOrder.GetOrdersRequest{
		UserId: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, ErrInternal
	}

	orders := make([]Order, 0, len(resp.Orders))
	for _, pbO := range resp.Orders {
		items := make([]OrderDish, 0, len(pbO.Items))
		for _, pbItem := range pbO.Items {
			items = append(items, OrderDish{
				DishID:      pbItem.DishId,
				DishName:    pbItem.DishName,
				Quantity:    pbItem.Quantity,
				Price:       pbItem.Price,
				OwnerUserID: pbItem.OwnerUserId,
			})
		}

		splits := make([]OrderSplit, 0, len(pbO.Splits))
		for _, pbSplit := range pbO.Splits {
			splits = append(splits, OrderSplit{
				SplitID:        pbSplit.SplitId,
				UserID:         pbSplit.UserId,
				BaseAmount:     pbSplit.BaseAmount,
				DiscountAmount: pbSplit.DiscountAmount,
				Amount:         pbSplit.Amount,
				Status:         pbSplit.Status,
			})
		}

		orders = append(orders, Order{
			PublicID:       pbO.PublicId,
			RestaurantName: pbO.RestaurantName,
			TotalCost:      pbO.TotalCost,
			Promocode:      pbO.AppliedPromocode,
			DiscountAmount: pbO.GetDiscountAmount(),
			Status:         pbO.Status,
			CreatedAt:      pbO.CreatedAt.AsTime(),
			Items:          items,
			Splits:         splits,
		})
	}

	return orders, nil
}

func (c *orderClient) PayForFriend(ctx context.Context, splitID string, payerID int64, paymentMethodID, idempotencyKey string) error {
	req := &pbOrder.PayForFriendRequest{
		SplitId:         splitID,
		PayerUserId:     payerID,
		PaymentMethodId: paymentMethodID,
		IdempotencyKey:  idempotencyKey,
	}

	_, err := c.client.PayForFriend(ctx, req)
	if err != nil {
		return ErrInternal
	}

	return nil
}

func (c *orderClient) CancelOrder(ctx context.Context, orderPublicID string, userID int64) error {
	_, err := c.client.CancelOrder(ctx, &pbOrder.CancelOrderRequest{
		OrderPublicId: orderPublicID,
		UserId:        userID,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				return ErrAddressNotFound
			case codes.PermissionDenied:
				return ErrInternal
			case codes.FailedPrecondition:
				return fmt.Errorf("cannot cancel: %s", st.Message())
			}
		}
		return fmt.Errorf("cancel order: %w", err)
	}
	return nil
}
