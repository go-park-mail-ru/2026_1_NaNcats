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
	DeliveryCost       int64  `json:"delivery_cost"`
	ServiceFee         int64  `json:"service_fee"`
	TotalCost          int64  `json:"total_cost"`
}

//easyjson:json
type CreateOrderResponse struct {
	OrderID         string `json:"order_id"`
	ConfirmationURL string `json:"confirmation_url,omitempty"`
}

//easyjson:json
type OrderHistoryResponse struct {
	OrderID            string `json:"order_id"`
	RestaurantName     string `json:"restaurant_name"`
	RestaurantImageURL string `json:"restaurant_image_url"`
	TotalCost          int64  `json:"total_cost"`
	Status             string `json:"status"`
	CreatedAt          string `json:"created_at"`
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
		l.Error("failed to get user_id from context", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	var req CreateOrderRequest
	if err = easyjson.UnmarshalFromReader(r.Body, &req); err != nil {
		l.Warn("failed to decode create order request", domain.Int("user_id", userID), domain.String("error", err.Error()))
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if req.AddressID == "" || req.RestaurantBranchID == 0 {
		l.Warn("create order request validation failed", domain.Int("user_id", userID))
		response.Error(w, http.StatusBadRequest, "Bad request")
		return
	}

	input := domain.CreateOrderInput{
		UserID:             userID,
		AddressPublicID:    req.AddressID,
		RestaurantBranchID: req.RestaurantBranchID,
		PaymentMethodID:    req.PaymentMethodID,
		DeliveryCost:       req.DeliveryCost,
		ServiceFee:         req.ServiceFee,
		TotalCost:          req.TotalCost,
	}

	orderPublicID, confirmationURL, err := h.orderUC.CreateOrder(ctx, userID, input)
	if err != nil {
		if errors.Is(err, domain.ErrCartIsEmpty) {
			l.Warn("order creation failed: cart is empty", domain.Int("user_id", userID))
			response.Error(w, http.StatusBadRequest, "Cart is empty")
		} else if errors.Is(err, domain.ErrAddressNotFound) {
			l.Warn("order creation failed: address not found", domain.Int("user_id", userID), domain.String("address_id", req.AddressID))
			response.Error(w, http.StatusNotFound, "Address not found")
		} else {
			l.Error("order creation failed unexpectedly", err, domain.Int("user_id", userID))
			response.Error(w, http.StatusInternalServerError, "Something went wrong")
		}
		return
	}

	l.Info("order created successfully",
		domain.Int("user_id", userID),
		domain.String("order_id", orderPublicID),
		domain.Any("payment_required", confirmationURL != ""),
	)

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
		l.Warn("unauthorized access to get orders")
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	orders, err := h.orderUC.GetOrders(ctx, userID)
	if err != nil {
		l.Error("failed to fetch user orders", err, domain.Int("user_id", userID))
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	resp := make([]OrderHistoryResponse, 0, len(orders))
	for _, o := range orders {
		resp = append(resp, OrderHistoryResponse{
			OrderID:            o.PublicID,
			RestaurantName:     o.PaymentMethodID,
			RestaurantImageURL: o.RestaurantLogoURL,
			TotalCost:          o.TotalCost,
			Status:             o.Status,
			CreatedAt:          o.CreatedAt.Format("02.01.2006"),
		})
	}

	l.Debug("successfully fetched user orders", domain.Int("user_id", userID), domain.Int("count", len(resp)))

	response.JSON(w, http.StatusOK, resp)
}
