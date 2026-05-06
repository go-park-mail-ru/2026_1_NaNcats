package grpc_client

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/usecase"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/order"
)

type orderClient struct {
	client pb.OrderServiceClient
}

func NewOrderClient(cl pb.OrderServiceClient) usecase.OrderClient {
	return &orderClient{
		client: cl,
	}
}

func (c *orderClient) UpdateOrderStatus(ctx context.Context, paymentID string, status string) error {
	_, err := c.client.UpdateOrderStatusByPaymentID(ctx, &pb.UpdateStatusRequest{
		YookassaPaymentId: paymentID,
		Status:            status,
	})

	return err
}
