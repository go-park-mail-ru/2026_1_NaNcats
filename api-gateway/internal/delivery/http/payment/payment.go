package payment

//go:generate easyjson $GOFILE

import (
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/paymentclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/api_clients/yookassa"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/response"
	"github.com/mailru/easyjson"
)

//easyjson:json
type PaymentMethodResponse struct {
	ID         string `json:"id" example:"pay-method-uuid"`
	CardType   string `json:"card_type" example:"Mir"`
	Last4      string `json:"last4" example:"6767"`
	IssuerName string `json:"issuer_name,omitempty" example:"Sber"`
	IsDefault  bool   `json:"is_default"`
}

//easyjson:json
type BindingResponse struct {
	ConfirmationURL string `json:"confirmation_url"`
}

type PaymentHandler struct {
	paymentClient paymentclient.PaymentClient
	logger        logger.Logger
}

func NewPaymentHandler(pc paymentclient.PaymentClient, l logger.Logger) *PaymentHandler {
	return &PaymentHandler{
		paymentClient: pc,
		logger:        l,
	}
}

// InitiateCardBinding godoc
// @Summary 		Инициализация привязки карты
// @Description		Создает запрос на привязку банковской карты пользователя и возвращает URL для подтверждения в ЮKassa
// @Tags			profile, payments
// @Produce			json
// @Success			200		{object}	BindingResponse "URL для подтверждения привязки"
// @Failure			401		{object}	map[string]string "Пользователь не авторизован"
// @Failure			500		{object}	map[string]string "Внутренняя ошибка сервера"
// @Router			/profile/cards/bind [post]
func (h *PaymentHandler) InitiateCardBinding(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		response.Error(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	confirmationURL, err := h.paymentClient.InitiateCardBinding(ctx, userID, idemKey)
	if err != nil {
		l.Error("failed to initiate card binding via grpc", err)
		response.Error(w, http.StatusInternalServerError, "failed to initiate payment method binding")
		return
	}

	response.JSON(w, http.StatusOK, BindingResponse{
		ConfirmationURL: confirmationURL,
	})
}

// GetUserCards godoc
// @Summary 		Получение сохраненных карт
// @Description		Возвращает список всех привязанных банковских карт пользователя
// @Tags			profile, payments
// @Produce			json
// @Success			200		{array}		PaymentMethodResponse
// @Failure			401		{object}	map[string]string "Пользователь не авторизован"
// @Failure			500		{object}	map[string]string "Внутренняя ошибка сервера"
// @Router			/profile/cards [get]
func (h *PaymentHandler) GetUserCards(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	cards, err := h.paymentClient.GetUserCards(ctx, userID)
	if err != nil {
		l.Error("failed to get users cards via grpc", err)
		response.Error(w, http.StatusInternalServerError, "failed to get payment methods")
		return
	}

	resp := make([]PaymentMethodResponse, 0, len(cards))
	for _, card := range cards {
		resp = append(resp, PaymentMethodResponse{
			ID:         card.ID,
			CardType:   card.CardType,
			Last4:      card.Last4,
			IssuerName: card.IssuerName,
			IsDefault:  card.IsDefault,
		})
	}

	response.JSON(w, http.StatusOK, resp)
}

// DeleteCard godoc
// @Summary 		Удаление карты
// @Description		Удаляет привязанную карту из профиля
// @Tags			profile, payments
// @Produce			json
// @Param			id		path		string		true	"ID карты"
// @Success			200
// @Failure			401		{object}	map[string]string "Пользователь не авторизован"
// @Failure			404		{object}	map[string]string "Карта не найдена"
// @Failure			500		{object}	map[string]string "Внутренняя ошибка сервера"
// @Router			/profile/cards/{id} [delete]
func (h *PaymentHandler) DeleteCard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		response.Error(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	cardID := r.PathValue("id")
	if cardID == "" {
		response.Error(w, http.StatusBadRequest, "Card ID is required")
		return
	}

	err := h.paymentClient.DeleteCard(ctx, userID, cardID, idemKey)
	if err != nil {
		if errors.Is(err, paymentclient.ErrPaymentMethodNotFound) {
			response.Error(w, http.StatusNotFound, "card not found")
			return
		}
		l.Error("failed to delete card via grpc", err)
		response.Error(w, http.StatusInternalServerError, "failed to delete payment method")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// SetDefaultCard godoc
// @Summary 		Выбор основной карты
// @Description		Устанавливает привязанную карту как основную (по умолчанию) для пользователя
// @Tags			profile, payments
// @Produce			json
// @Param			id		path		string		true	"ID карты"
// @Success			200
// @Failure			401		{object}	map[string]string "Пользователь не авторизован"
// @Failure			404		{object}	map[string]string "Карта не найдена"
// @Failure			500		{object}	map[string]string "Внутренняя ошибка сервера"
func (h *PaymentHandler) SetDefaultCard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		response.Error(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	cardID := r.PathValue("id")
	if cardID == "" {
		response.Error(w, http.StatusBadRequest, "Card ID is required")
		return
	}

	err := h.paymentClient.SetDefaultCard(ctx, userID, cardID, idemKey)
	if err != nil {
		if errors.Is(err, paymentclient.ErrPaymentMethodNotFound) {
			response.Error(w, http.StatusNotFound, "card not found")
			return
		}
		l.Error("failed to set default card via grpc", err)
		response.Error(w, http.StatusInternalServerError, "failed to set default payment method")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "default card updated"})
}

// YookassaWebhook godoc
// @Summary 		Вебхук ЮKassa
// @Description		Обрабатывает асинхронные уведомления от ЮKassa (например, об успешной привязке платежного метода)
// @Tags			payments, webhooks
// @Accept			json
// @Produce			json
// @Param			notification	body		yookassa.WebhookNotification	true	"Данные уведомления от ЮKassa"
// @Success			200
// @Failure			400		{object}	map[string]string "Неверный формат данных"
// @Router			/webhooks/yookassa [post]
func (h *PaymentHandler) YookassaWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	var notification yookassa.WebhookNotification
	if err := easyjson.UnmarshalFromReader(r.Body, &notification); err != nil {
		l.Warn("invalid webhook payload from yookassa", logger.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "invalid payload")
		return
	}
	defer r.Body.Close()

	switch notification.Event {
	case "payment_method.active":
		var methodObj yookassa.WebhookPaymentMethodObject
		if err := easyjson.Unmarshal(notification.Object, &methodObj); err != nil {
			l.Error("failed to parse payment_method object from webhook", err)
			w.WriteHeader(http.StatusOK) // Отдаем 200, чтобы ЮКасса не спамила ретраями
			return
		}

		if methodObj.Card == nil {
			l.Warn("card object is nil in webhook")
			w.WriteHeader(http.StatusOK)
			return
		}

		cardInfo := paymentclient.CardInfo{
			First6:      methodObj.Card.First6,
			Last4:       methodObj.Card.Last4,
			ExpiryMonth: methodObj.Card.ExpiryMonth,
			ExpiryYear:  methodObj.Card.ExpiryYear,
			CardType:    methodObj.Card.CardType,
			IssuerName:  methodObj.Card.IssuerName,
		}

		err := h.paymentClient.ProcessPaymentMethodWebhook(ctx, methodObj.ID, methodObj.Status, methodObj.Type, methodObj.Saved, cardInfo)
		if err != nil {
			l.Error("grpc error processing payment method webhook", err)
		}

	case "payment.succeeded", "payment.canceled":
		var paymentObj yookassa.WebhookPaymentObject
		if err := easyjson.Unmarshal(notification.Object, &paymentObj); err != nil {
			l.Error("failed to parse payment object from webhook", err)
			w.WriteHeader(http.StatusOK)
			return
		}

		err := h.paymentClient.ProcessPaymentWebhook(ctx, paymentObj.ID, paymentObj.Status)
		if err != nil {
			l.Error("grpc error processing payment webhook", err)
		}
	}

	w.WriteHeader(http.StatusOK)
}
