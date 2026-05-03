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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderClient interface {
	UpdateOrderStatus(ctx context.Context, paymentID string, status string) error
}

//go:generate mockgen -destination=mocks/payment_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/usecase PaymentUseCase
//go:generate gowrap gen -i PaymentUseCase -t ../../../../shared/templates/tracing.tmpl -o payment_tracing_mw.go -v TracerName=payment-service
type PaymentUseCase interface {
	CreatePayment(ctx context.Context, amount int64, paymentMethodID, idempotencyKey string) (string, string, error)
	InitiateCardBinding(ctx context.Context, userID int64, idempotencyKey string) (string, error)
	GetUserCards(ctx context.Context, userID int64) ([]domain.PaymentMethod, error)
	SetDefaultCard(ctx context.Context, cardID string, userID int64, idempotencyKey string) error
	DeleteCard(ctx context.Context, cardID string, userID int64, idempotencyKey string) error

	ProcessPaymentMethodWebhook(ctx context.Context, paymentMethod *yookassa.WebhookPaymentMethodObject) error
	ProcessPaymentWebhook(ctx context.Context, payment *yookassa.WebhookPaymentObject) error
	RefreshPaymentStatus(ctx context.Context, paymentID string) (string, error)
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

func (p *paymentUseCase) CreatePayment(ctx context.Context, amount int64, paymentMethodID, idempotencyKey string) (paymentID string, confirmationURL string, err error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("payment.amount_raw", amount),
		attribute.String("payment.method_id", paymentMethodID),
		attribute.String("payment.idempotency_key", idempotencyKey),
	)

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
		// Confirmation указываем всегда: для новой карты это redirect на форму YooKassa,
		// для сохранённой — fallback на случай, когда YooKassa внезапно требует 3DS;
		// без return_url пользователь застрянет на странице YooKassa после подтверждения.
		Confirmation: &yookassa.CreatePaymentRequestConfirmation{
			Type:      "redirect",
			ReturnURL: p.returnURL,
		},
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
			return "", "", errutil.Internal("internal payment configuration error", err)

		default:
			return "", "", errutil.Internal("unexpected error from yookassa", err)
		}
	}

	paymentID = paymentResponse.ID
	span.SetAttributes(attribute.String("payment.external_id", paymentID))

	if paymentResponse.Confirmation != nil && paymentResponse.Confirmation.Type == "redirect" {
		confirmationURL = paymentResponse.Confirmation.ConfirmationURL
	}

	return paymentResponse.ID, confirmationURL, nil
}

func (p *paymentUseCase) InitiateCardBinding(ctx context.Context, userID int64, idempotencyKey string) (confirmationURL string, err error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("user.id", userID),
		attribute.String("binding.idempotency_key", idempotencyKey),
	)

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
			return "", errutil.Internal("payment service auth failure", err)
		}
		return "", errutil.Internal("failed to contact yookassa", err)
	}

	span.SetAttributes(attribute.String("binding.external_id", resp.ID))

	if resp.Confirmation == nil || resp.Confirmation.ConfirmationURL == "" {
		return "", errutil.Internal("empty confirmation url from yookassa", errors.New("malformed provider response"))
	}

	confirmationURL = resp.Confirmation.ConfirmationURL

	err = p.cacheRepo.SetPendingBinding(ctx, resp.ID, userID, 15*time.Minute)
	if err != nil {
		return "", errutil.Internal("failed to save pending binding state", err)
	}

	return resp.Confirmation.ConfirmationURL, nil
}

func (p *paymentUseCase) GetUserCards(ctx context.Context, userID int64) ([]domain.PaymentMethod, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int64("user.id", userID))

	methods, err := p.paymentRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errutil.Internal("failed to retrieve cards from database", err)
	}

	span.SetAttributes(attribute.Int("payment_methods.count", len(methods)))
	return methods, nil
}

func (p *paymentUseCase) SetDefaultCard(ctx context.Context, cardID string, userID int64, idempotencyKey string) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("user.id", userID),
		attribute.String("card.external_id", cardID),
	)

	err := p.paymentRepo.SetDefault(ctx, cardID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentMethodNotFound) {
			return err
		}
		return errutil.Internal("failed to set default card", err)
	}
	return nil
}

func (p *paymentUseCase) DeleteCard(ctx context.Context, cardID string, userID int64, idempotencyKey string) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("user.id", userID),
		attribute.String("card.external_id", cardID),
	)

	err := p.paymentRepo.Delete(ctx, cardID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentMethodNotFound) {
			return err
		}
		return errutil.Internal("failed to delete card", err)
	}

	return nil
}

func (p *paymentUseCase) ProcessPaymentMethodWebhook(ctx context.Context, pm *yookassa.WebhookPaymentMethodObject) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("payment_method.external_id", pm.ID),
		attribute.String("payment_method.status", pm.Status),
	)

	if !pm.Saved || pm.Status != "active" || pm.Card == nil {
		span.AddEvent("ignoring_webhook_unsupported_state")
		return nil
	}

	userID, err := p.cacheRepo.GetUserIDByPaymentID(ctx, pm.ID)
	if err != nil {
		return errutil.Wrap("WEBHOOK_CACHE_MISS", "payment context expired or not found in cache", err, codes.NotFound)
	}
	span.SetAttributes(attribute.Int64("user.id", userID))

	span.SetAttributes(
		attribute.String("card.type", pm.Card.CardType),
		attribute.String("card.issuer", pm.Card.IssuerName),
	)

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
			span.AddEvent("payment_method_already_exists")
			return nil
		}
		return errutil.Internal("failed to save payment method to db", err)
	}

	err = p.cacheRepo.DeletePendingBinding(ctx, pm.ID)
	if err != nil {
		span.AddEvent("cache_cleanup_failed")
		p.logger.WithContext(ctx).Warn("failed to delete pending binding from cache",
			logger.String("payment_id", pm.ID),
			logger.Err(err),
		)
	}

	return nil
}

// RefreshPaymentStatus тянет актуальный статус из YooKassa REST и применяет
// его как обычный webhook (через ProcessPaymentWebhook). Используется когда
// YooKassa-вебхук не доходит до нашего сервера (например, dev на localhost).
func (p *paymentUseCase) RefreshPaymentStatus(ctx context.Context, paymentID string) (string, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("payment.external_id", paymentID))

	if paymentID == "" {
		return "", errutil.Wrap("PAYMENT_INVALID_DATA", "payment id is empty", errors.New("empty payment id"), codes.InvalidArgument)
	}

	resp, err := p.yookassaClient.GetPayment(ctx, paymentID)
	if err != nil {
		if errors.Is(err, yookassa.ErrNotFound) {
			return "", errutil.Wrap("PAYMENT_NOT_FOUND", "payment not found in yookassa", err, codes.NotFound)
		}
		return "", errutil.Internal("failed to fetch payment from yookassa", err)
	}

	span.SetAttributes(attribute.String("payment.status", resp.Status))

	// Применяем тот же путь, что и для веб-хука
	if err := p.ProcessPaymentWebhook(ctx, &yookassa.WebhookPaymentObject{
		ID:     resp.ID,
		Status: resp.Status,
	}); err != nil {
		return resp.Status, err
	}

	return resp.Status, nil
}

func (p *paymentUseCase) ProcessPaymentWebhook(ctx context.Context, payment *yookassa.WebhookPaymentObject) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("payment.external_id", payment.ID),
		attribute.String("payment.status", payment.Status),
	)

	if payment.Status != "succeeded" && payment.Status != "canceled" {
		span.AddEvent("ignoring_webhook_unsupported_status")
		return nil
	}

	err := p.orderClient.UpdateOrderStatus(ctx, payment.ID, payment.Status)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			span.AddEvent("order_not_found_for_payment")
			p.logger.WithContext(ctx).Error("order not found for incoming payment", err,
				logger.String("payment_id", payment.ID),
			)
			return nil
		}
		return errutil.Internal("failed to notify order service", err)
	}

	return nil
}
