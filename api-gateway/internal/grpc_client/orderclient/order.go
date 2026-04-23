package orderclient

import (
	"context"
	"errors"
	"time"

	pbOrder "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/order"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrCartIsEmpty     = errors.New("cart is empty")
	ErrAddressNotFound = errors.New("address not found")
	ErrInternal        = errors.New("internal server error")
)

type CreateOrderInput struct {
	AddressPublicID    string
	RestaurantBranchID int64
	RestaurantBrandID  int64
	PaymentMethodID    string
	DeliveryCost       int64
	ServiceFee         int64
}

type Order struct {
	PublicID          string
	RestaurantName    string
	RestaurantLogoURL string
	TotalCost         int64
	Status            string
	CreatedAt         time.Time
}

type OrderClient interface {
	CreateOrder(ctx context.Context, userID int64, input CreateOrderInput, idempotencyKey string) (string, string, error)
	GetOrders(ctx context.Context, userID int64) ([]Order, error)
}

type orderClient struct {
	client pbOrder.OrderServiceClient
}

func NewOrderClient(cl pbOrder.OrderServiceClient) OrderClient {
	return &orderClient{client: cl}
}

func (c *orderClient) CreateOrder(ctx context.Context, userID int64, input CreateOrderInput, idempotencyKey string) (string, string, error) {
	req := &pbOrder.CreateOrderRequest{
		UserId:             userID,
		AddressPublicId:    input.AddressPublicID,
		RestaurantBranchId: input.RestaurantBranchID,
		RestaurantBrandId:  input.RestaurantBrandID,
		PaymentMethodId:    input.PaymentMethodID,
		DeliveryCost:       input.DeliveryCost,
		ServiceFee:         input.ServiceFee,
		IdempotencyKey:     idempotencyKey,
	}

	resp, err := c.client.CreateOrder(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				return "", "", ErrAddressNotFound
			case codes.FailedPrecondition, codes.InvalidArgument:
				return "", "", ErrCartIsEmpty
			}
		}
		return "", "", ErrInternal
	}

	return resp.OrderPublicId, resp.ConfirmationUrl, nil
}

func (c *orderClient) GetOrders(ctx context.Context, userID int64) ([]Order, error) {
	resp, err := c.client.GetOrders(ctx, &pbOrder.GetOrdersRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, ErrInternal
	}

	orders := make([]Order, 0, len(resp.Orders))
	for _, pbO := range resp.Orders {
		orders = append(orders, Order{
			PublicID:          pbO.PublicId,
			RestaurantName:    pbO.RestaurantName,
			RestaurantLogoURL: pbO.RestaurantLogoUrl,
			TotalCost:         pbO.TotalCost,
			Status:            pbO.Status,
			CreatedAt:         pbO.CreatedAt.AsTime(),
		})
	}

	return orders, nil
}
