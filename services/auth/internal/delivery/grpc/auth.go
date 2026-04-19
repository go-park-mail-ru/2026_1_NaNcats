package grpc

import (
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/auth/internal/usecase"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/auth"
)

type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
	usecase usecase.AuthUseCase
}

func NewAuthHandler(auc usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		usecase: auc,
	}
}
