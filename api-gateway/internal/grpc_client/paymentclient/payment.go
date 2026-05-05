package paymentclient

import (
	"context"
	"errors"

	pbPayment "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/payment"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrPaymentMethodNotFound = errors.New("payment method not found")
	ErrInternal              = errors.New("internal server error")
)

//go:generate mockgen -destination=../../../../shared/proto/payment/mocks/payment_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/payment PaymentServiceClient
type PaymentMethod struct {
	ID         string
	CardType   string
	Last4      string
	IssuerName string
	IsDefault  bool
}

type CardInfo struct {
	First6      string
	Last4       string
	ExpiryMonth string
	ExpiryYear  string
	CardType    string
	IssuerName  string
}

type PaymentClient interface {
	InitiateCardBinding(ctx context.Context, userID int64, idempotencyKey string) (string, error)
	GetUserCards(ctx context.Context, userID int64) ([]PaymentMethod, error)
	SetDefaultCard(ctx context.Context, userID int64, cardID string, idempotencyKey string) error
	DeleteCard(ctx context.Context, userID int64, cardID string, idempotencyKey string) error
	ProcessPaymentMethodWebhook(ctx context.Context, id, status, pType string, saved bool, card CardInfo) error
	ProcessPaymentWebhook(ctx context.Context, id, status string) error
	RefreshPaymentStatus(ctx context.Context, paymentID string) (string, error)
}

type paymentClient struct {
	client pbPayment.PaymentServiceClient
}

func NewPaymentClient(cl pbPayment.PaymentServiceClient) PaymentClient {
	return &paymentClient{client: cl}
}

func (c *paymentClient) InitiateCardBinding(ctx context.Context, userID int64, idempotencyKey string) (string, error) {
	resp, err := c.client.InitiateCardBinding(ctx, &pbPayment.InitiateCardBindingRequest{
		UserId:         userID,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return "", ErrInternal
	}
	return resp.ConfirmationUrl, nil
}

func (c *paymentClient) GetUserCards(ctx context.Context, userID int64) ([]PaymentMethod, error) {
	resp, err := c.client.GetUserCards(ctx, &pbPayment.GetUserCardsRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, ErrInternal
	}

	cards := make([]PaymentMethod, 0, len(resp.Cards))
	for _, pbCard := range resp.Cards {
		cards = append(cards, PaymentMethod{
			ID:         pbCard.ExternalId,
			CardType:   pbCard.CardType,
			Last4:      pbCard.Last4,
			IssuerName: pbCard.IssuerName,
			IsDefault:  pbCard.IsDefault,
		})
	}
	return cards, nil
}

func (c *paymentClient) SetDefaultCard(ctx context.Context, userID int64, cardID string, idempotencyKey string) error {
	_, err := c.client.SetDefaultCard(ctx, &pbPayment.ChangeCardRequest{
		UserId:         userID,
		CardId:         cardID,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return ErrPaymentMethodNotFound
		}
		return ErrInternal
	}
	return nil
}

func (c *paymentClient) DeleteCard(ctx context.Context, userID int64, cardID string, idempotencyKey string) error {
	_, err := c.client.DeleteCard(ctx, &pbPayment.ChangeCardRequest{
		UserId:         userID,
		CardId:         cardID,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return ErrPaymentMethodNotFound
		}
		return ErrInternal
	}
	return nil
}

func (c *paymentClient) ProcessPaymentMethodWebhook(ctx context.Context, id, status, pType string, saved bool, card CardInfo) error {
	_, err := c.client.ProcessPaymentMethodWebhook(ctx, &pbPayment.ProcessPaymentMethodWebhookRequest{
		Id:     id,
		Status: status,
		Saved:  saved,
		Type:   pType,
		Card: &pbPayment.CardInfo{
			First6:      card.First6,
			Last4:       card.Last4,
			ExpiryMonth: card.ExpiryMonth,
			ExpiryYear:  card.ExpiryYear,
			CardType:    card.CardType,
			IssuerName:  card.IssuerName,
		},
	})
	if err != nil {
		return ErrInternal
	}
	return nil
}

func (c *paymentClient) ProcessPaymentWebhook(ctx context.Context, id, status string) error {
	_, err := c.client.ProcessPaymentWebhook(ctx, &pbPayment.ProcessPaymentWebhookRequest{
		Id:     id,
		Status: status,
	})
	if err != nil {
		return ErrInternal
	}
	return nil
}

func (c *paymentClient) RefreshPaymentStatus(ctx context.Context, paymentID string) (string, error) {
	resp, err := c.client.RefreshPaymentStatus(ctx, &pbPayment.RefreshPaymentStatusRequest{
		YookassaPaymentId: paymentID,
	})
	if err != nil {
		return "", ErrInternal
	}
	return resp.Status, nil
}
