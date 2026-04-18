package grpc

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/grpcutil"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/user"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Маппер из доменной структуры в Protobuf
func mapDomainToPBUser(u domain.User) *pb.User {
	return &pb.User{
		Id:           u.ID,
		Name:         u.Name,
		Email:        u.Email,
		Role:         u.Role,
		AvatarUrl:    u.AvatarURL,
		PasswordHash: u.PasswordHash,
	}
}

func mapDomainToPBClientProfile(u domain.ClientProfile) *pb.ClientProfile {
	pbProfile := &pb.ClientProfile{
		AccountId:    u.AccountID,
		BonusBalance: u.BonusBalance,
		StreakCount:  int32(u.StreakCount),
	}

	if u.BonusCategoryID != nil {
		id := *u.BonusCategoryID
		pbProfile.BonusCategoryId = &id
	}

	if u.BonusCategoryExpiresAt != nil {
		pbProfile.BonusCategoryExpiresAt = timestamppb.New(*u.BonusCategoryExpiresAt)
	}

	if u.BonusExpiresAt != nil {
		pbProfile.BonusExpiresAt = timestamppb.New(*u.BonusExpiresAt)
	}

	if u.LastOrderDate != nil {
		pbProfile.LastOrderDate = timestamppb.New(*u.LastOrderDate)
	}

	if u.PremiumExpiresAt != nil {
		pbProfile.PremiumExpiresAt = timestamppb.New(*u.PremiumExpiresAt)
	}

	return pbProfile
}

type UserHandler struct {
	pb.UnimplementedUserServiceServer
	userUC   usecase.UserUseCase
	clientUC usecase.ClientProfileUseCase
}

func NewUserHandler(uuc usecase.UserUseCase, cpuc usecase.ClientProfileUseCase) *UserHandler {
	return &UserHandler{
		userUC:   uuc,
		clientUC: cpuc,
	}
}

func (h *UserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	userToCreate := domain.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: req.PasswordHash,
	}

	createdUserID, err := h.userUC.Create(ctx, userToCreate, req.IdempotencyKey)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &pb.CreateUserResponse{
		UserId: createdUserID,
	}, nil
}

func (h *UserHandler) CreateClientProfile(ctx context.Context, req *pb.CreateClientProfileRequest) (*emptypb.Empty, error) {
	err := h.clientUC.CreateProfile(ctx, req.UserId, req.IdempotencyKey)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &emptypb.Empty{}, nil
}

func (h *UserHandler) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*emptypb.Empty, error) {
	err := h.userUC.UpdateProfile(ctx, req.UserId, req.Name, req.Email, req.IdempotencyKey)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &emptypb.Empty{}, nil
}

func (h *UserHandler) UpdateAvatar(ctx context.Context, req *pb.UpdateAvatarRequest) (*pb.UpdateAvatarResponse, error) {
	avatarURL, err := h.userUC.UpdateAvatar(ctx, req.UserId, req.ImageData, req.IdempotencyKey)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &pb.UpdateAvatarResponse{
		AvatarUrl: avatarURL,
	}, nil
}

func (h *UserHandler) DeleteAvatar(ctx context.Context, req *pb.DeleteAvatarRequest) (*pb.DeleteAvatarResponse, error) {
	avatarURL, err := h.userUC.DeleteAvatar(ctx, req.UserId, req.IdempotencyKey)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &pb.DeleteAvatarResponse{
		DefaultAvatarUrl: avatarURL,
	}, nil
}

func (h *UserHandler) GetByID(ctx context.Context, req *pb.GetUserByIDRequest) (*pb.GetUserResponse, error) {
	user, err := h.userUC.GetByID(ctx, req.UserId)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &pb.GetUserResponse{
		User: mapDomainToPBUser(user),
	}, nil
}

func (h *UserHandler) GetByEmail(ctx context.Context, req *pb.GetUserByEmailRequest) (*pb.GetUserResponse, error) {
	user, err := h.userUC.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &pb.GetUserResponse{
		User: mapDomainToPBUser(user),
	}, nil
}

func (h *UserHandler) CheckUserExists(ctx context.Context, req *pb.CheckUserExistsRequest) (*pb.CheckUserExistsResponse, error) {
	isExists, err := h.userUC.Check(ctx, req.UserId)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &pb.CheckUserExistsResponse{
		Exists: isExists,
	}, nil
}

func (h *UserHandler) GetUserProfile(ctx context.Context, req *pb.GetUserProfileRequest) (*pb.GetUserProfileResponse, error) {
	user, err := h.userUC.GetByID(ctx, req.UserId)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}
	client, err := h.clientUC.GetByAccountID(ctx, req.UserId)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &pb.GetUserProfileResponse{
		User:    mapDomainToPBUser(user),
		Profile: mapDomainToPBClientProfile(client),
	}, nil
}
