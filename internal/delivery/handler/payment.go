package handler

import (
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/api_clients/yookassa"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/response"
	"github.com/mailru/easyjson"
)

type paymentHandler struct {
	paymentUC usecase.PaymentUseCase
	logger    domain.Logger
}

func NewPaymentHandler(puc usecase.PaymentUseCase, logger domain.Logger) *paymentHandler {
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

	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	confirmationURL, err := h.paymentUC.InitiateCardBinding(ctx, userID)
	if err != nil {
		h.logger.Error("failed to initiate card binding", err, map[string]any{"user_id": userID})
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
// @Success			200		{array}		domain.PaymentMethod
// @Failure			401		{object}	map[string]string "Пользователь не авторизован"
// @Failure			500		{object}	map[string]string "Внутренняя ошибка сервера"
// @Router			/profile/cards [get]
func (h *paymentHandler) GetUserCards(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value(middleware.UserIDKey).(int)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cards, err := h.paymentUC.GetUserCards(ctx, userID)
	if err != nil {
		h.logger.Error("failed to get users cards", err, map[string]any{
			"user_id": userID,
		})
		response.Error(w, http.StatusInternalServerError, "failed to get payment methods")
		return
	}

	if cards == nil {
		cards = make([]domain.PaymentMethod, 0)
	}

	response.JSON(w, http.StatusOK, cards)
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

	userID, ok := ctx.Value(middleware.UserIDKey).(int)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cardID := r.PathValue("id")
	err := h.paymentUC.DeleteCard(ctx, cardID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentMethodNotFound) {
			response.Error(w, http.StatusNotFound, "card not found")
			return
		}
		h.logger.Error("failed to delete card", err, map[string]any{
			"user_id": userID,
			"card_id": cardID,
		})
		response.Error(w, http.StatusInternalServerError, "failed to delete payment method")
		return
	}

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

	userID, ok := ctx.Value(middleware.UserIDKey).(int)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	cardID := r.PathValue("id")
	err := h.paymentUC.SetDefaultCard(ctx, cardID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentMethodNotFound) {
			response.Error(w, http.StatusNotFound, "card not found")
			return
		}
		h.logger.Error("failed to set default card", err, map[string]any{
			"user_id": userID,
			"card_id": cardID,
		})
		response.Error(w, http.StatusInternalServerError, "failed to set default payment method")
		return
	}

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

	var notification yookassa.WebhookNotification
	if err := easyjson.UnmarshalFromReader(r.Body, &notification); err != nil {
		h.logger.Warn("invalid webhook payload from yookassa", map[string]any{"error": err.Error()})
		response.Error(w, http.StatusBadRequest, "invalid payload")
		return
	}
	defer r.Body.Close()

	if notification.Event == "payment_method.active" {
		err := h.paymentUC.ProcessWebhook(ctx, &notification.Object)
		if err != nil {
			h.logger.Error("failed to process yookassa webhook", err, map[string]any{
				"payment_id": notification.Object.ID,
			})
		}
	}

	w.WriteHeader(http.StatusOK)
}
