package grpc_client

import (
	"context"

	authDomain "github.com/go-park-mail-ru/2026_1_NaNcats/services/auth/internal/domain"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/user"
)

type userClient struct {
	client pb.UserServiceClient
}

func NewUserClient(uc pb.UserServiceClient) UserClient {
	return &userClient{
		client: uc,
	}
}

func (c *userClient) GetUserByEmail(ctx context.Context, email string) (authDomain.User, error) {
	resp, err := c.client.GetByEmail(ctx, &pb.GetUserByEmailRequest{Email: email})
	if err != nil {
		return authDomain.User{}, err
	}

	return authDomain.User{
		ID:           resp.User.Id,
		Email:        resp.User.Email,
		PasswordHash: resp.User.PasswordHash,
		Role:         resp.User.Role,
	}, nil
}
