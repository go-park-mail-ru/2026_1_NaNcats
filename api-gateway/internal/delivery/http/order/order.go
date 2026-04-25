package order

//go:generate easyjson $GOFILE

import (
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/orderclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/request"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/response"

	wsManager "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/delivery/websocket"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

//easyjson:json
type CreateOrderRequest struct {
	AddressID          string `json:"address_id"`
	RestaurantBranchID int64  `json:"branch_id"`
	RestaurantBrandID  int64  `json:"brand_id"`
	PaymentMethodID    string `json:"payment_method_id,omitempty"`
	DeliveryCost       int64  `json:"delivery_cost"`
	ServiceFee         int64  `json:"service_fee"`
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

type OrderHandler struct {
	orderClient orderclient.OrderClient
	wsManager   *wsManager.WsManager
	logger      logger.Logger
}

func NewOrderHandler(oc orderclient.OrderClient, wsm *wsManager.WsManager, l logger.Logger) *OrderHandler {
	return &OrderHandler{
		orderClient: oc,
		wsManager:   wsm,
		logger:      l,
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
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
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

	var req CreateOrderRequest
	if err := request.JSON(r, &req); err != nil {
		l.Warn("failed to decode create order request", logger.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.AddressID == "" || req.RestaurantBranchID == 0 {
		response.Error(w, http.StatusBadRequest, "address_id and branch_id are required")
		return
	}

	input := orderclient.CreateOrderInput{
		AddressPublicID:    req.AddressID,
		RestaurantBranchID: req.RestaurantBranchID,
		RestaurantBrandID:  req.RestaurantBrandID,
		PaymentMethodID:    req.PaymentMethodID,
		DeliveryCost:       req.DeliveryCost,
		ServiceFee:         req.ServiceFee,
	}

	orderPublicID, confirmationURL, err := h.orderClient.CreateOrder(ctx, userID, input, idemKey)
	if err != nil {
		if errors.Is(err, orderclient.ErrCartIsEmpty) {
			response.Error(w, http.StatusBadRequest, "Cart is empty")
			return
		} else if errors.Is(err, orderclient.ErrAddressNotFound) {
			response.Error(w, http.StatusNotFound, "Address not found")
			return
		}

		l.Error("order creation failed via grpc", err)
		response.Error(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	l.Info("order created successfully",
		logger.String("order_id", orderPublicID),
		logger.Any("payment_required", confirmationURL != ""),
	)

	response.JSON(w, http.StatusOK, CreateOrderResponse{
		OrderID:         orderPublicID,
		ConfirmationURL: confirmationURL,
	})
}

func (h *OrderHandler) GetMyOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	orders, err := h.orderClient.GetOrders(ctx, userID)
	if err != nil {
		l.Error("failed to fetch user orders via grpc", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	resp := make([]OrderHistoryResponse, 0, len(orders))
	for _, o := range orders {
		resp = append(resp, OrderHistoryResponse{
			OrderID:            o.PublicID,
			RestaurantName:     o.RestaurantName,
			RestaurantImageURL: o.RestaurantLogoURL,
			TotalCost:          o.TotalCost,
			Status:             o.Status,
			CreatedAt:          o.CreatedAt.Format("02.01.2006"),
		})
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *OrderHandler) TrackOrderWS(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue("id")
	if orderID == "" {
		response.Error(w, http.StatusBadRequest, "order_id is required")
		return
	}

	// TODO: проверить проверку на принадлежность заказа конкретному юзеру

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("failed to upgrade to websocket", err)
		return
	}

	h.wsManager.AddConnection(orderID, conn)
}
