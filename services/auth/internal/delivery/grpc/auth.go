package grpc

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/auth/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/auth/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/grpcutil"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/auth"
	"github.com/google/uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func mapDomainToPBSession(s domain.Session) *pb.Session {
	return &pb.Session{
		UserId:    s.UserID,
		UserAgent: s.UserAgent,
		Role:      s.Role,
		ExpiresAt: timestamppb.New(s.ExpiresAt),
		Id:        s.ID.String(),
	}
}

type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
	usecase usecase.AuthUseCase
}

func NewAuthHandler(auc usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		usecase: auc,
	}
}

func (h *AuthHandler) IssueSession(ctx context.Context, req *pb.IssueSessionRequest) (*pb.AuthResponse, error) {
	session, err := h.usecase.IssueSession(ctx, req.UserId, req.Role, req.UserAgent)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &pb.AuthResponse{
		Session: mapDomainToPBSession(session),
	}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	session, err := h.usecase.Login(ctx, req.Email, req.Password, req.UserAgent)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &pb.AuthResponse{
		Session: mapDomainToPBSession(session),
	}, nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*emptypb.Empty, error) {
	sessionID, err := uuid.Parse(req.SessionId)
	if err != nil {
		wrappedErr := errutil.Wrap("INVALID_ID_SESSION_FORMAT", "invalid session id format", err, codes.InvalidArgument)
		return nil, grpcutil.ToGRPCError(wrappedErr)
	}

	err = h.usecase.Logout(ctx, sessionID)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &emptypb.Empty{}, nil
}

func (h *AuthHandler) CheckSession(ctx context.Context, req *pb.CheckSessionRequest) (*pb.CheckSessionResponse, error) {
	sessionID, err := uuid.Parse(req.SessionId)
	if err != nil {
		wrappedErr := errutil.Wrap("INVALID_ID_SESSION_FORMAT", "invalid session id format", err, codes.InvalidArgument)
		return nil, grpcutil.ToGRPCError(wrappedErr)
	}

	userID, userRole, err := h.usecase.CheckUserSession(ctx, sessionID)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &pb.CheckSessionResponse{
		UserId: userID,
		Role:   userRole,
	}, nil
}

func (h *AuthHandler) GetCSRF(ctx context.Context, req *pb.CSRFRequest) (*pb.CSRFResponse, error) {
	sessionID, err := uuid.Parse(req.SessionId)
	if err != nil {
		wrappedErr := errutil.Wrap("INVALID_ID_SESSION_FORMAT", "invalid session id format", err, codes.InvalidArgument)
		return nil, grpcutil.ToGRPCError(wrappedErr)
	}

	token, err := h.usecase.GetCSRFBySessionID(ctx, sessionID)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &pb.CSRFResponse{
		Token: token,
	}, nil
}

func (h *AuthHandler) SetCSRF(ctx context.Context, req *pb.CSRFRequest) (*pb.CSRFResponse, error) {
	sessionID, err := uuid.Parse(req.SessionId)
	if err != nil {
		wrappedErr := errutil.Wrap("INVALID_ID_SESSION_FORMAT", "invalid session id format", err, codes.InvalidArgument)
		return nil, grpcutil.ToGRPCError(wrappedErr)
	}

	token, err := h.usecase.SetCSRFForUser(ctx, sessionID)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &pb.CSRFResponse{
		Token: token,
	}, nil
}
