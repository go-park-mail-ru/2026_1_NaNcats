package grpc_client

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/usecase"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/user"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type userClient struct {
	client pb.UserServiceClient
}

func NewUserClient(cl pb.UserServiceClient) usecase.UserClient {
	return &userClient{client: cl}
}

func (c *userClient) OnOrderPaid(ctx context.Context, userID, restaurantID int64, orderPublicID string, paidAt time.Time) error {
	_, err := c.client.OnOrderPaid(ctx, &pb.OnOrderPaidRequest{
		UserId:        userID,
		OrderPublicId: orderPublicID,
		RestaurantId:  restaurantID,
		PaidAt:        timestamppb.New(paidAt),
	})
	return err
}
