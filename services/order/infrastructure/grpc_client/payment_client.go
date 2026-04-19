package grpc_client

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/usecase"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/payment"
)

type paymentClient struct {
	client pb.PaymentServiceClient
}

func NewPaymentClient(cl pb.PaymentServiceClient) usecase.PaymentClient {
	return &paymentClient{
		client: cl,
	}
}

func (c *paymentClient) CreatePayment(ctx context.Context, amount int64, paymentMethodID string, idempotencyKey string) (string, string, error) {
	resp, err := c.client.CreatePayment(ctx, &pb.CreatePaymentRequest{
		Amount:          amount,
		PaymentMethodId: paymentMethodID,
		IdempotencyKey:  idempotencyKey,
	})
	if err != nil {
		return "", "", err
	}

	return resp.PaymentId, resp.ConfirmationUrl, nil
}
