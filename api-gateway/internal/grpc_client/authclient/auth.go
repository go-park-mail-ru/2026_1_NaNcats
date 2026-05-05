package authclient

import (
	"context"
	"errors"

	pbAuth "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrSessionNotFound    = errors.New("session not found or expired")
	ErrInternal           = errors.New("internal server error")
)

//go:generate mockgen -destination=../../../../shared/proto/auth/mocks/auth_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/auth AuthServiceClient
type AuthClient interface {
	IssueSession(ctx context.Context, userID int64, role, userAgent string) (*pbAuth.Session, error)
	Login(ctx context.Context, email, password, userAgent string) (*pbAuth.Session, error)
	Logout(ctx context.Context, sessionID string) error
	CheckSession(ctx context.Context, sessionID string) (int64, string, error)
	SetCSRF(ctx context.Context, sessionID string) (string, error)
	GetCSRF(ctx context.Context, sessionID string) (string, error)
}

type authClient struct {
	client pbAuth.AuthServiceClient
}

func NewAuthClient(cl pbAuth.AuthServiceClient) AuthClient {
	return &authClient{
		client: cl,
	}
}

func (c *authClient) IssueSession(ctx context.Context, userID int64, role, userAgent string) (*pbAuth.Session, error) {
	resp, err := c.client.IssueSession(ctx, &pbAuth.IssueSessionRequest{
		UserId:    userID,
		Role:      role,
		UserAgent: userAgent,
	})
	if err != nil {
		return nil, ErrInternal
	}

	return resp.Session, nil
}

func (c *authClient) Login(ctx context.Context, email, password, userAgent string) (*pbAuth.Session, error) {
	resp, err := c.client.Login(ctx, &pbAuth.LoginRequest{
		Email:     email,
		Password:  password,
		UserAgent: userAgent,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.Unauthenticated {
			return nil, ErrInvalidCredentials
		}
		return nil, ErrInternal
	}

	return resp.Session, nil
}

func (c *authClient) Logout(ctx context.Context, sessionID string) error {
	_, err := c.client.Logout(ctx, &pbAuth.LogoutRequest{
		SessionId: sessionID,
	})
	return err // Нам не так важно, если сессии уже нет, поэтому отдаем как есть
}

func (c *authClient) CheckSession(ctx context.Context, sessionID string) (int64, string, error) {
	resp, err := c.client.CheckSession(ctx, &pbAuth.CheckSessionRequest{
		SessionId: sessionID,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && (st.Code() == codes.NotFound || st.Code() == codes.Unauthenticated) {
			return 0, "", ErrSessionNotFound
		}
		return 0, "", ErrInternal
	}
	return resp.UserId, resp.Role, nil
}

func (c *authClient) SetCSRF(ctx context.Context, sessionID string) (string, error) {
	resp, err := c.client.SetCSRF(ctx, &pbAuth.CSRFRequest{
		SessionId: sessionID,
	})
	if err != nil {
		return "", ErrInternal
	}
	return resp.Token, nil
}

func (c *authClient) GetCSRF(ctx context.Context, sessionID string) (string, error) {
	resp, err := c.client.GetCSRF(ctx, &pbAuth.CSRFRequest{
		SessionId: sessionID,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return "", ErrSessionNotFound
		}
		return "", ErrInternal
	}
	return resp.Token, nil
}
