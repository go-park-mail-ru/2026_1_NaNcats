package userclient

import (
	"context"
	"errors"
	"time"

	pbUser "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrEmailAlreadyExists  = errors.New("email already exists")
	ErrInvalidArgument     = errors.New("invalid argument or empty update data")
	ErrInternal            = errors.New("internal server error")
	ErrWheelCooldownActive = errors.New("wheel cooldown active")
)

//go:generate mockgen -destination=mocks/user_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/userclient UserClient

type Achievement struct {
	ID          int64
	Code        string
	Title       string
	Description string
	Icon        string
	SortOrder   int32
}

type UserAchievement struct {
	AchievementID int64
	AwardedAt     time.Time
}

type UserClient interface {
	CreateUser(ctx context.Context, name, email, password, idempotencyKey string) (int64, error)
	GetByID(ctx context.Context, userID int64) (*pbUser.User, error)
	GetUserProfile(ctx context.Context, userID int64) (*pbUser.User, *pbUser.ClientProfile, error)
	UpdateProfile(ctx context.Context, userID int64, name, email *string, idempotencyKey string) error
	UpdateAvatar(ctx context.Context, userID int64, fileBytes []byte, idempotencyKey string) (string, error)
	DeleteAvatar(ctx context.Context, userID int64, idempotencyKey string) (string, error)
	UpdateRole(ctx context.Context, userID int64, newRole string, idempotencyKey string) error
	GetUsersByIDs(ctx context.Context, userIDs []int64) (map[int64]*pbUser.User, error)
	ResolvePublicID(ctx context.Context, publicID string) (int64, error)
	ListAchievements(ctx context.Context) ([]Achievement, error)
	GetUserAchievements(ctx context.Context, userID int64) ([]UserAchievement, error)
	ClaimWheelSpin(ctx context.Context, userID int64) error
	ResetWheelSpinCooldown(ctx context.Context, userID int64) error
	OnWheelSpin(ctx context.Context, userID int64, wonCode string) error
	ActivateStreakFreeze(ctx context.Context, userID int64) error
	IncrementStreak(ctx context.Context, userID int64) error
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

func (c *userClient) GetUsersByIDs(ctx context.Context, userIDs []int64) (map[int64]*pbUser.User, error) {
	resp, err := c.client.GetUsersByIDs(ctx, &pbUser.GetUsersByIDsRequest{
		UserIds: userIDs,
	})
	if err != nil {
		return nil, ErrInternal
	}

	if resp.Users == nil {
		return make(map[int64]*pbUser.User), nil
	}
	return resp.Users, nil
}

func (c *userClient) ListAchievements(ctx context.Context) ([]Achievement, error) {
	resp, err := c.client.ListAchievements(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, ErrInternal
	}
	out := make([]Achievement, 0, len(resp.Achievements))
	for _, a := range resp.Achievements {
		out = append(out, Achievement{
			ID:          a.Id,
			Code:        a.Code,
			Title:       a.Title,
			Description: a.Description,
			Icon:        a.Icon,
			SortOrder:   a.SortOrder,
		})
	}
	return out, nil
}

func (c *userClient) GetUserAchievements(ctx context.Context, userID int64) ([]UserAchievement, error) {
	resp, err := c.client.GetUserAchievements(ctx, &pbUser.GetUserAchievementsRequest{UserId: userID})
	if err != nil {
		return nil, ErrInternal
	}
	out := make([]UserAchievement, 0, len(resp.Achievements))
	for _, ua := range resp.Achievements {
		var awardedAt time.Time
		if ua.AwardedAt != nil {
			awardedAt = ua.AwardedAt.AsTime()
		}
		out = append(out, UserAchievement{
			AchievementID: ua.AchievementId,
			AwardedAt:     awardedAt,
		})
	}
	return out, nil
}

func (c *userClient) ResolvePublicID(ctx context.Context, publicID string) (int64, error) {
	resp, err := c.client.ResolvePublicID(ctx, &pbUser.ResolvePublicIDRequest{
		PublicId: publicID,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return 0, ErrUserNotFound
		}
		return 0, ErrInternal
	}
	return resp.UserId, nil
}

func (c *userClient) ActivateStreakFreeze(ctx context.Context, userID int64) error {
	_, err := c.client.ActivateStreakFreeze(ctx, &pbUser.ActivateStreakFreezeRequest{
		UserId: userID,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return ErrUserNotFound
		}
		return ErrInternal
	}
	return nil
}

func (c *userClient) IncrementStreak(ctx context.Context, userID int64) error {
	_, err := c.client.IncrementStreak(ctx, &pbUser.IncrementStreakRequest{
		UserId: userID,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return ErrUserNotFound
		}
		return ErrInternal
	}
	return nil
}

func (c *userClient) OnWheelSpin(ctx context.Context, userID int64, wonCode string) error {
	var wonCodePtr *string
	if wonCode != "" {
		wonCodePtr = &wonCode
	}

	_, err := c.client.OnWheelSpin(ctx, &pbUser.OnWheelSpinRequest{
		UserId:             userID,
		WonAchievementCode: wonCodePtr,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return ErrUserNotFound
		}
		return ErrInternal
	}
	return nil
}

func (c *userClient) ClaimWheelSpin(ctx context.Context, userID int64) error {
	_, err := c.client.ClaimWheelSpin(ctx, &pbUser.ClaimWheelSpinRequest{
		UserId: userID,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.FailedPrecondition:
				return ErrWheelCooldownActive
			case codes.NotFound:
				return ErrUserNotFound
			}
		}
		return ErrInternal
	}
	return nil
}

func (c *userClient) ResetWheelSpinCooldown(ctx context.Context, userID int64) error {
	_, err := c.client.ResetWheelSpinCooldown(ctx, &pbUser.ResetWheelSpinCooldownRequest{
		UserId: userID,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return ErrUserNotFound
		}
		return ErrInternal
	}
	return nil
}
