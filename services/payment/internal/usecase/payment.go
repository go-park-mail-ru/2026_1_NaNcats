package usecase

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/api_clients/yookassa"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	logger         logger.Logger
}

func NewPaymentUseCase(pr repository.PaymentRepository, cr repository.PaymentCacheRepository, oc OrderClient, yc *yookassa.Client, returnURL string, l logger.Logger) PaymentUseCase {
	return &paymentUseCase{
		paymentRepo:    pr,
		cacheRepo:      cr,
		orderClient:    oc,
		yookassaClient: yc,
		returnURL:      returnURL,
		logger:         l,
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
		switch {
		case errors.Is(err, yookassa.ErrBadRequest):
			return "", "", errutil.Wrap("PAYMENT_INVALID_DATA", "invalid payment data provided to provider", err, codes.InvalidArgument)

		case errors.Is(err, yookassa.ErrNotFound):
			return "", "", errutil.Wrap("PAYMENT_METHOD_NOT_FOUND", "provided payment method was not found in yookassa", err, codes.NotFound)

		case errors.Is(err, yookassa.ErrUnauthorized):
			return "", "", errutil.Wrap("PAYMENT_CONFIG_ERROR", "internal payment configuration error", err, codes.Internal)

		default:
			return "", "", errutil.Wrap("PAYMENT_PROVIDER_ERROR", "unexpected error from yookassa", err, codes.Internal)
		}
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
		if errors.Is(err, yookassa.ErrBadRequest) {
			return "", errutil.Wrap("BINDING_INVALID_CONFIG", "yookassa rejected binding request", err, codes.InvalidArgument)
		}
		if errors.Is(err, yookassa.ErrUnauthorized) {
			return "", errutil.Wrap("BINDING_AUTH_ERROR", "payment service auth failure", err, codes.Internal)
		}

		return "", errutil.Wrap("BINDING_PROVIDER_ERROR", "failed to contact yookassa", err, codes.Internal)
	}

	if resp.Confirmation == nil || resp.Confirmation.ConfirmationURL == "" {
		return "", errutil.New("BINDING_MALFORMED_RESPONSE", "empty confirmation url from yookassa", codes.Internal)
	}

	err = p.cacheRepo.SetPendingBinding(ctx, resp.ID, userID, 15*time.Minute)
	if err != nil {
		return "", errutil.Wrap("INTERNAL_SERVER_ERROR", "failed to save pending binding state", err, codes.Internal)
	}

	return resp.Confirmation.ConfirmationURL, nil
}

func (p *paymentUseCase) GetUserCards(ctx context.Context, userID int64) ([]domain.PaymentMethod, error) {
	methods, err := p.paymentRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errutil.Wrap("PAYMENT_METHODS_FETCH_ERROR", "failed to retrieve cards from database", err, codes.Internal)
	}
	return methods, nil
}

func (p *paymentUseCase) SetDefaultCard(ctx context.Context, cardID string, userID int64, idempotencyKey string) error {
	err := p.paymentRepo.SetDefault(ctx, cardID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentMethodNotFound) {
			return err
		}
		return errutil.Wrap(
			"SET_DEFAULT_CARD_FAILED",
			"failed to set default card",
			err,
			codes.Internal,
		)
	}
	return nil
}

func (p *paymentUseCase) DeleteCard(ctx context.Context, cardID string, userID int64, idempotencyKey string) error {
	err := p.paymentRepo.Delete(ctx, cardID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentMethodNotFound) {
			return err
		}
		return errutil.Wrap("CARD_DELETE_FAILED", "failed to delete card due to internal error", err, codes.Internal)
	}

	return nil
}

func (p *paymentUseCase) ProcessPaymentMethodWebhook(ctx context.Context, pm *yookassa.WebhookPaymentMethodObject) error {
	if !pm.Saved || pm.Status != "active" || pm.Card == nil {
		return nil
	}

	userID, err := p.cacheRepo.GetUserIDByPaymentID(ctx, pm.ID)
	if err != nil {
		return errutil.Wrap("WEBHOOK_CACHE_MISS", "payment context expired or not found in cache", err, codes.NotFound)
	}

	domainPaymentMethod := domain.PaymentMethod{
		UserID:      userID,
		ExternalID:  pm.ID,
		First6:      pm.Card.First6,
		Last4:       pm.Card.Last4,
		ExpiryMonth: pm.Card.ExpiryMonth,
		ExpiryYear:  pm.Card.ExpiryYear,
		CardType:    pm.Card.CardType,
		IssuerName:  pm.Card.IssuerName,
		IsDefault:   false,
	}

	_, err = p.paymentRepo.Create(ctx, domainPaymentMethod, pm.ID)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentMethodAlreadyExists) {
			return nil
		}
		return errutil.Wrap("WEBHOOK_DATABASE_ERROR", "failed to save payment method to db", err, codes.Internal)
	}

	_ = p.cacheRepo.DeletePendingBinding(ctx, pm.ID)
	if err != nil {
		p.logger.WithContext(ctx).Warn("failed to delete pending binding from cache",
			logger.String("payment_id", pm.ID),
			logger.Err(err),
		)
	}

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
		st, ok := status.FromError(err)

		if ok && st.Code() == codes.NotFound {
			p.logger.WithContext(ctx).Error("order not found for incoming payment", err,
				logger.String("payment_id", payment.ID),
			)
			return nil
		}
		return errutil.Wrap(
			"ORDER_SERVICE_UPDATE_FAILED",
			"failed to notify order service about payment status",
			err,
			codes.Unavailable,
		)

	}

	return nil
}
