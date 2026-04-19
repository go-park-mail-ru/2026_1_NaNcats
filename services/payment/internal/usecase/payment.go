package payment

import (
	"context"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/api_clients/yookassa"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"google.golang.org/grpc/codes"
)

type OrderClient interface {
	UpdateOrderStatus(ctx context.Context, paymentID string, status string) error
}

//go:generate mockgen -destination=mocks/payment_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/payment PaymentUseCase
type PaymentUseCase interface {
	CreatePayment(ctx context.Context, amount int64, paymentMethodID, idempotencyKey string) (string, string, error)
	InitiateCardBinding(ctx context.Context, userID int64, idempotencyKey string) (string, error)
	GetUserCards(ctx context.Context, userID int64) ([]domain.PaymentMethod, error)
	SetDefaultCard(ctx context.Context, cardID string, userID int64, idempotencyKey string) error
	DeleteCard(ctx context.Context, cardID string, userID int64, idempotencyKey string) error

	ProcessPaymentMethodWebhook(ctx context.Context, paymentMethod *yookassa.WebhookPaymentMethodObject) error
	ProcessPaymentWebhook(ctx context.Context, payment *yookassa.WebhookPaymentObject) error
}

type paymentUseCase struct {
	paymentRepo    repository.PaymentRepository
	cacheRepo      repository.PaymentCacheRepository
	orderClient    OrderClient
	yookassaClient *yookassa.Client
	returnURL      string
}

func NewPaymentUseCase(pr repository.PaymentRepository, cr repository.PaymentCacheRepository, oc OrderClient, yc *yookassa.Client, returnURL string) PaymentUseCase {
	return &paymentUseCase{
		paymentRepo:    pr,
		cacheRepo:      cr,
		orderClient:    oc,
		yookassaClient: yc,
		returnURL:      returnURL,
	}
}

func (p *paymentUseCase) CreatePayment(ctx context.Context, amount int64, paymentMethodID, idempotencyKey string) (string, string, error) {
	rubles := amount / 1_000_000
	kopecks := (amount%1_000_000)/10_000 + 100
	value := strconv.FormatInt(rubles, 10) + "." + strconv.FormatInt(kopecks, 10)[1:]

	paymentRequest := yookassa.CreatePaymentRequest{
		Amount: yookassa.CreatePaymentRequestAmount{
			Value:    value,
			Currency: "RUB",
		},
		Capture:           true,
		SavePaymentMethod: false,
	}

	paymentRequest.Confirmation = &yookassa.CreatePaymentRequestConfirmation{
		Type:      "redirect",
		ReturnURL: p.returnURL,
	}

	if paymentMethodID != "" {
		paymentRequest.PaymentMethodID = paymentMethodID
	}

	paymentResponse, err := p.yookassaClient.CreatePayment(ctx, paymentRequest, idempotencyKey)
	if err != nil {
		return "", "", errutil.Wrap("failed to create payment in yookassa", err, codes.Internal)
	}

	var confirmationURL string
	if paymentResponse.Confirmation != nil && paymentResponse.Confirmation.Type == "redirect" {
		confirmationURL = paymentResponse.Confirmation.ConfirmationURL
	}

	return paymentResponse.ID, confirmationURL, nil
}

func (p *paymentUseCase) InitiateCardBinding(ctx context.Context, userID int64, idempotencyKey string) (string, error) {
	req := yookassa.CreatePaymentMethodRequest{
		Type: "bank_card",
		Confirmation: &yookassa.PaymentMethodRequestConfirmation{
			Type:      "redirect",
			ReturnURL: p.returnURL,
		},
	}

	resp, err := p.yookassaClient.CreatePaymentMethod(ctx, req, idempotencyKey)
	if err != nil {
		return "", errutil.Wrap("failed to initiate card binding", err, codes.Internal)
	}

	if resp.Confirmation == nil || resp.Confirmation.ConfirmationURL == "" {
		return "", errutil.New("empty confirmation url from payment provider", codes.Internal)
	}

	err = p.cacheRepo.SetPendingBinding(ctx, resp.ID, userID, 15*time.Minute)
	if err != nil {
		return "", errutil.Wrap("failed to save pending binding in cache", err, codes.Internal)
	}

	return resp.Confirmation.ConfirmationURL, nil
}

func (p *paymentUseCase) GetUserCards(ctx context.Context, userID int64) ([]domain.PaymentMethod, error) {
	methods, err := p.paymentRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errutil.Wrap("failed to get user cards", err, codes.Internal)
	}
	return methods, nil
}

func (p *paymentUseCase) SetDefaultCard(ctx context.Context, cardID string, userID int64, idempotencyKey string) error {
	err := p.paymentRepo.SetDefault(ctx, cardID, userID)
	if err != nil {
		return errutil.Wrap("failed to set default card", err, codes.Internal)
	}
	return nil
}

func (p *paymentUseCase) DeleteCard(ctx context.Context, cardID string, userID int64, idempotencyKey string) error {
	err := p.paymentRepo.Delete(ctx, cardID, userID)
	if err != nil {
		return errutil.Wrap("failed to delete card", err, codes.Internal)
	}
	return nil
}

func (p *paymentUseCase) ProcessPaymentMethodWebhook(ctx context.Context, pm *yookassa.WebhookPaymentMethodObject) error {
	if !pm.Saved || pm.Status != "active" || pm.Card == nil {
		return nil
	}

	userID, err := p.cacheRepo.GetUserIDByPaymentID(ctx, pm.ID)
	if err != nil {
		return errutil.Wrap("failed to find user_id for payment_method in cache", err, codes.NotFound)
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
		return errutil.Wrap("failed to save payment method to db", err, codes.Internal)
	}

	_ = p.cacheRepo.DeletePendingBinding(ctx, pm.ID)

	return nil
}

func (p *paymentUseCase) ProcessPaymentWebhook(ctx context.Context, payment *yookassa.WebhookPaymentObject) error {
	if payment.Status != "succeeded" && payment.Status != "canceled" {
		return nil
	}

	var newStatus string
	switch payment.Status {
	case "succeeded":
		newStatus = "finished"
	case "canceled":
		newStatus = "canceled"
	default:
		return nil
	}

	err := p.orderClient.UpdateOrderStatus(ctx, payment.ID, newStatus)
	if err != nil {
		return errutil.Wrap("failed to notify order service about payment status", err, codes.Internal)
	}

	return nil
}
