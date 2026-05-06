package userclient

import (
	"context"
	"errors"

	pbUser "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidArgument    = errors.New("invalid argument or empty update data")
	ErrInternal           = errors.New("internal server error")
)

//go:generate mockgen -destination=mocks/user_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/userclient UserClient
type UserClient interface {
	CreateUser(ctx context.Context, name, email, password, idempotencyKey string) (int64, error)
	GetByID(ctx context.Context, userID int64) (*pbUser.User, error)
	GetUserProfile(ctx context.Context, userID int64) (*pbUser.User, *pbUser.ClientProfile, error)
	UpdateProfile(ctx context.Context, userID int64, name, email *string, idempotencyKey string) error
	UpdateAvatar(ctx context.Context, userID int64, fileBytes []byte, idempotencyKey string) (string, error)
	DeleteAvatar(ctx context.Context, userID int64, idempotencyKey string) (string, error)
	UpdateRole(ctx context.Context, userID int64, newRole string, idempotencyKey string) error
}

type userClient struct {
	client pbUser.UserServiceClient
}

func NewUserClient(cl pbUser.UserServiceClient) UserClient {
	return &userClient{client: cl}
}

func (c *userClient) CreateUser(ctx context.Context, name, email, password, idempotencyKey string) (int64, error) {
	resp, err := c.client.CreateUser(ctx, &pbUser.CreateUserRequest{
		Name:           name,
		Email:          email,
		Password:       password,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			return 0, ErrEmailAlreadyExists
		}
		return 0, ErrInternal
	}
	return resp.UserId, nil
}

func (c *userClient) GetByID(ctx context.Context, userID int64) (*pbUser.User, error) {
	resp, err := c.client.GetByID(ctx, &pbUser.GetUserByIDRequest{UserId: userID})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return nil, ErrUserNotFound
		}
		return nil, ErrInternal
	}
	return resp.User, nil
}

func (c *userClient) GetUserProfile(ctx context.Context, userID int64) (*pbUser.User, *pbUser.ClientProfile, error) {
	resp, err := c.client.GetUserProfile(ctx, &pbUser.GetUserProfileRequest{UserId: userID})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return nil, nil, ErrUserNotFound
		}
		return nil, nil, ErrInternal
	}
	return resp.User, resp.Profile, nil
}

func (c *userClient) UpdateProfile(ctx context.Context, userID int64, name, email *string, idempotencyKey string) error {
	req := &pbUser.UpdateProfileRequest{
		UserId:         userID,
		Name:           name,
		Email:          email,
		IdempotencyKey: idempotencyKey,
	}

	_, err := c.client.UpdateProfile(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.AlreadyExists:
				return ErrEmailAlreadyExists
			case codes.InvalidArgument:
				return ErrInvalidArgument
			}
		}
		return ErrInternal
	}
	return nil
}

func (c *userClient) UpdateAvatar(ctx context.Context, userID int64, fileBytes []byte, idempotencyKey string) (string, error) {
	resp, err := c.client.UpdateAvatar(ctx, &pbUser.UpdateAvatarRequest{
		UserId:         userID,
		ImageData:      fileBytes,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.InvalidArgument {
			return "", ErrInvalidArgument
		}
		return "", ErrInternal
	}
	return resp.AvatarUrl, nil
}

func (c *userClient) DeleteAvatar(ctx context.Context, userID int64, idempotencyKey string) (string, error) {
	resp, err := c.client.DeleteAvatar(ctx, &pbUser.DeleteAvatarRequest{
		UserId:         userID,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return "", ErrInternal
	}
	return resp.DefaultAvatarUrl, nil
}

func (c *userClient) UpdateRole(ctx context.Context, userID int64, newRole string, idempotencyKey string) error {
	_, err := c.client.UpdateUserRole(ctx, &pbUser.UpdateUserRoleRequest{
		UserId:         userID,
		NewRole:        newRole,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.InvalidArgument:
				return ErrInvalidArgument
			case codes.NotFound:
				return ErrUserNotFound
			}
		}
		return ErrInternal
	}
	return nil
}
