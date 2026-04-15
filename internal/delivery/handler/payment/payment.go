package payment

//go:generate easyjson $GOFILE

import (
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	payment "github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/payment"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/api_clients/yookassa"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/response"
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

type paymentHandler struct {
	paymentUC payment.PaymentUseCase
	logger    logger.Logger
}

func NewPaymentHandler(puc payment.PaymentUseCase, logger logger.Logger) *paymentHandler {
	return &paymentHandler{
		paymentUC: puc,
		logger:    logger,
	}
}

//easyjson:json
type BindingResponse struct {
	ConfirmationURL string `json:"confirmation_url"`
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
func (h *paymentHandler) InitiateCardBinding(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		l.Error("failed to get user_id from context", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	confirmationURL, err := h.paymentUC.InitiateCardBinding(ctx, userID)
	if err != nil {
		l.Error("failed to initiate card binding", err, logger.Int("user_id", userID))
		response.Error(w, http.StatusInternalServerError, "failed to initiate payment method binding")
		return
	}

	l.Debug("card binding initiated", logger.Int("user_id", userID))

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
func (h *paymentHandler) GetUserCards(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := ctx.Value(middleware.UserIDKey).(int)
	if !ok {
		l.Warn("unauthorized access to get cards")
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cards, err := h.paymentUC.GetUserCards(ctx, userID)
	if err != nil {
		l.Error("failed to get users cards", err, logger.Int("user_id", userID))
		response.Error(w, http.StatusInternalServerError, "failed to get payment methods")
		return
	}

	if cards == nil {
		cards = make([]domain.PaymentMethod, 0)
	}

	l.Debug("successfully fetched user cards", logger.Int("user_id", userID), logger.Int("count", len(cards)))

	resp := make([]PaymentMethodResponse, 0, len(cards))
	for _, card := range cards {
		resp = append(resp, PaymentMethodResponse{
			ID:         card.ExternalID,
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
func (h *paymentHandler) DeleteCard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := ctx.Value(middleware.UserIDKey).(int)
	if !ok {
		l.Warn("unauthorized access to delete card")
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cardID := r.PathValue("id")
	err := h.paymentUC.DeleteCard(ctx, cardID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentMethodNotFound) {
			l.Warn("card not found for deletion", logger.Int("user_id", userID), logger.String("card_id", cardID))
			response.Error(w, http.StatusNotFound, "card not found")
			return
		}
		l.Error("failed to delete card", err, logger.Int("user_id", userID), logger.String("card_id", cardID))
		response.Error(w, http.StatusInternalServerError, "failed to delete payment method")
		return
	}

	l.Info("card deleted successfully", logger.Int("user_id", userID), logger.String("card_id", cardID))
	w.WriteHeader(http.StatusOK)
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
func (h *paymentHandler) SetDefaultCard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := ctx.Value(middleware.UserIDKey).(int)
	if !ok {
		l.Warn("unauthorized access to set default card")
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cardID := r.PathValue("id")
	err := h.paymentUC.SetDefaultCard(ctx, cardID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentMethodNotFound) {
			l.Warn("card not found to set as default", logger.Int("user_id", userID), logger.String("card_id", cardID))
			response.Error(w, http.StatusNotFound, "card not found")
			return
		}
		l.Error("failed to set default card", err, logger.Int("user_id", userID), logger.String("card_id", cardID))
		response.Error(w, http.StatusInternalServerError, "failed to set default payment method")
		return
	}

	l.Info("default card updated", logger.Int("user_id", userID), logger.String("card_id", cardID))
	w.WriteHeader(http.StatusOK)
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
func (h *paymentHandler) YookassaWebhook(w http.ResponseWriter, r *http.Request) {
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
			w.WriteHeader(http.StatusOK)
			return
		}

		err := h.paymentUC.ProcessPaymentMethodWebhook(ctx, &methodObj)
		if err != nil {
			l.Error("failed to process payment method webhook", err, logger.String("payment_id", methodObj.ID))
		} else {
			l.Info("payment method activated via webhook", logger.String("payment_id", methodObj.ID))
		}
	case "payment.succeeded", "payment.canceled":
		var paymentObj yookassa.WebhookPaymentObject
		if err := easyjson.Unmarshal(notification.Object, &paymentObj); err != nil {
			l.Error("failed to parse payment object from webhook", err)
			w.WriteHeader(http.StatusOK)
			return
		}

		err := h.paymentUC.ProcessPaymentWebhook(ctx, &paymentObj)
		if err != nil {
			l.Error("failed to process payment webhook", err, logger.String("payment_id", paymentObj.ID), logger.String("event", notification.Event))
		} else {
			l.Info("payment status updated via webhook", logger.String("payment_id", paymentObj.ID), logger.String("event", notification.Event))
		}
	}

	w.WriteHeader(http.StatusOK)
}
