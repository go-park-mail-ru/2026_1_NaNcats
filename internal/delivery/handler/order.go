package handler

//go:generate easyjson $GOFILE

import (
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/response"
	"github.com/mailru/easyjson"
)

//easyjson:json
type CreateOrderRequest struct {
	AddressID          string `json:"address_id"`
	RestaurantBranchID int    `json:"branch_id"`
	PaymentMethodID    string `json:"payment_method_id,omitempty"`
}

//easyjson:json
type CreateOrderResponse struct {
	OrderID         string `json:"order_id"`
	ConfirmationURL string `json:"confirmation_url,omitempty"`
}

//easyjson:json
type OrderHistoryResponse struct {
	OrderID        string    `json:"order_id"`
	RestaurantName string    `json:"restaurant_name"`
	TotalCost      int64     `json:"total_cost"`
	Status         string    `json:"status"`
	CreatedAt      string    `json:"created_at"`
}

type orderHandler struct {
	orderUC usecase.OrderUseCase
	logger  domain.Logger
}

func NewOrderHandler(ouc usecase.OrderUseCase, l domain.Logger) *orderHandler {
	return &orderHandler{
		orderUC: ouc,
		logger:  l,
	}
}

// CreateOrder godoc
// @Summary      Создать заказ
// @Description  Создает заказ на основе корзины, возвращает ID заказа и ссылку на оплату YooKassa (при необходимости)
// @Tags         order
// @Accept       json
// @Produce      json
// @Param        input  body      CreateOrderRequest  true  "Данные для оформления заказа"
// @Success      200    {object}  CreateOrderResponse "Заказ успешно создан"
// @Failure      400    {object}  map[string]string   "Bad request или пустая корзина"
// @Failure      401    {object}  map[string]string   "Unauthorized"
// @Failure      404    {object}  map[string]string   "Указанный адрес не найден"
// @Failure      500    {object}  map[string]string   "Internal server error"
// @Security     ApiKeyAuth
// @Router       /api/orders [post]
func (h *orderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		l.Error("auth middleware missed userID in route", err, nil)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	var req CreateOrderRequest
	if err = easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		l.Error("failed to parse request body", err, nil)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if req.AddressID == "" || req.RestaurantBranchID == 0 {
		l.Warn("create order bad request", nil)
		response.Error(w, http.StatusBadRequest, "Bad request")
		return
	}

	inpit := domain.CreateOrderInput{
		UserID:             userID,
		AddressPublicID:    req.AddressID,
		RestaurantBranchID: req.RestaurantBranchID,
		PaymentMethodID:    req.PaymentMethodID,
	}

	orderPublicID, confirmationURL, err := h.orderUC.CreateOrder(ctx, userID, inpit)
	if err != nil {
		if errors.Is(err, domain.ErrCartIsEmpty) {
			response.Error(w, http.StatusBadRequest, "Cart is empty")
		} else if errors.Is(err, domain.ErrAddressNotFound) {
			response.Error(w, http.StatusNotFound, "Address not found")
		} else {
			response.Error(w, http.StatusInternalServerError, "Something went wrong")
		}
		return
	}

	response.JSON(w, http.StatusOK, CreateOrderResponse{
		OrderID:         orderPublicID,
		ConfirmationURL: confirmationURL,
	})
}

func (h *orderHandler) GetMyOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	orders, err := h.orderUC.GetOrders(ctx, userID)
	if err != nil {
		l.Error("failed to get user orders", err, map[string]any{"user_id": userID})
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	resp := make([]OrderHistoryResponse, 0, len(orders))
	for _, o := range orders {
		resp = append(resp, OrderHistoryResponse{
			OrderID:        o.PublicID,
			RestaurantName: o.PaymentMethodID,
			TotalCost:      o.TotalCost,
			Status:         o.Status,
			CreatedAt:      o.CreatedAt.Format("02.01.2006"),
		})
	}

	response.JSON(w, http.StatusOK, resp)
}
