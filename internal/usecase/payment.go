package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/api_clients/yookassa"
)

type PaymentUseCase interface {
	InitiateCardBinding(ctx context.Context, userID int) (string, error)
	GetUserCards(ctx context.Context, userID int) ([]domain.PaymentMethod, error)
	SetDefaultCard(ctx context.Context, cardID string, userID int) error
	DeleteCard(ctx context.Context, cardID string, userID int) error
	ProcessWebhook(ctx context.Context, paymentMethod *yookassa.WebhookPaymentMethodObject) error
}

type paymentUseCase struct {
	paymentRepo    repository.PaymentRepository
	cacheRepo      repository.PaymentCacheRepository
	yookassaClient *yookassa.Client
	returnURL      string
}

func NewPaymentUseCase(pr repository.PaymentRepository, cr repository.PaymentCacheRepository, yc *yookassa.Client, returnURL string) PaymentUseCase {
	return &paymentUseCase{
		paymentRepo:    pr,
		cacheRepo:      cr,
		yookassaClient: yc,
		returnURL:      returnURL,
	}
}

func (p *paymentUseCase) InitiateCardBinding(ctx context.Context, userID int) (string, error) {
	req := yookassa.CreatePaymentMethodRequest{
		Type: "bank_card",
		Confirmation: &yookassa.PaymentMethodRequestConfirmation{
			Type:      "redirect",
			ReturnURL: p.returnURL,
		},
	}

	resp, err := p.yookassaClient.CreatePaymentMethod(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to initiate card binding: %w", err)
	}

	if resp.Confirmation == nil || resp.Confirmation.ConfirmationURL == "" {
		return "", domain.ErrYookassaConfirmationURL
	}

	err = p.cacheRepo.SetPendingBinding(ctx, resp.ID, userID, 15*time.Minute)
	if err != nil {
		return "", fmt.Errorf("failed to save pending binding in cache: %w", err)
	}

	return resp.Confirmation.ConfirmationURL, nil
}

func (p *paymentUseCase) GetUserCards(ctx context.Context, userID int) ([]domain.PaymentMethod, error) {
	return p.paymentRepo.GetByUserID(ctx, userID)
}

func (p *paymentUseCase) SetDefaultCard(ctx context.Context, cardID string, userID int) error {
	return p.paymentRepo.SetDefault(ctx, cardID, userID)
}

func (p *paymentUseCase) DeleteCard(ctx context.Context, cardID string, userID int) error {
	return p.paymentRepo.Delete(ctx, cardID, userID)
}

func (p *paymentUseCase) ProcessWebhook(ctx context.Context, pm *yookassa.WebhookPaymentMethodObject) error {
	if !pm.Saved || pm.Status != "active" || pm.Card == nil {
		return nil
	}

	userID, err := p.cacheRepo.GetUserIDByPaymentID(ctx, pm.ID)
	if err != nil {
		return fmt.Errorf("failed to find user_id for payment_method %s: %w", pm.ID, err)
	}

	issuer := ""
	if pm.Card.IssuerName != "" {
		issuer = pm.Card.IssuerName
	}

	domainPaymentMethod := domain.PaymentMethod{
		UserID:     userID,
		ExternalID: pm.ID,
		CardType:   pm.Card.CardType,
		Last4:      pm.Card.Last4,
		IssuerName: issuer,
		IsDefault:  false,
	}

	_, err = p.paymentRepo.Create(ctx, domainPaymentMethod)
	if err != nil {
		return fmt.Errorf("failed to save payment method to db: %w", err)
	}

	_ = p.cacheRepo.DeletePendingBinding(ctx, pm.ID)

	return nil
}
