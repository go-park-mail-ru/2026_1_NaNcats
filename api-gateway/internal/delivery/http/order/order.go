package order

//go:generate easyjson $GOFILE

import (
	"errors"
	"net/http"

	wsManager "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/delivery/websocket"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/orderclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/paymentclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/restaurantclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/request"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/response"

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

	PayForAll    bool            `json:"pay_for_all"`
	PayerMapping map[int64]int64 `json:"payer_mapping,omitempty"`
}

//easyjson:json
type CreateOrderResponse struct {
	OrderID string `json:"order_id"`
}

//easyjson:json
type OrderDishDTO struct {
	DishID      int64  `json:"dish_id"`
	Name        string `json:"name"`
	ImageURL    string `json:"image_url"`
	Quantity    int32  `json:"quantity"`
	Price       int64  `json:"price"`
	OwnerUserID *int64 `json:"owner_user_id,omitempty"`
}

//easyjson:json
type OrderSplitDTO struct {
	SplitID string `json:"split_id"`
	UserID  int64  `json:"user_id"`
	Amount  int64  `json:"amount"`
	Status  string `json:"status"`
}

//easyjson:json
type OrderHistoryResponse struct {
	OrderID            string          `json:"order_id"`
	RestaurantName     string          `json:"restaurant_name"`
	RestaurantImageURL string          `json:"restaurant_image_url"`
	TotalCost          int64           `json:"total_cost"`
	Status             string          `json:"status"`
	CreatedAt          string          `json:"created_at"`
	Items              []OrderDishDTO  `json:"items"`
	Splits             []OrderSplitDTO `json:"splits"`
}

//easyjson:json
type PayForFriendRequest struct {
	PaymentMethodID string `json:"payment_method_id"`
}

type OrderHandler struct {
	orderClient      orderclient.OrderClient
	paymentClient    paymentclient.PaymentClient
	restaurantClient restaurantclient.RestaurantClient
	wsManager        *wsManager.WsManager
	logger           logger.Logger
}

func NewOrderHandler(oc orderclient.OrderClient, pc paymentclient.PaymentClient, rc restaurantclient.RestaurantClient, wsm *wsManager.WsManager, l logger.Logger) *OrderHandler {
	return &OrderHandler{
		orderClient:      oc,
		paymentClient:    pc,
		restaurantClient: rc,
		wsManager:        wsm,
		logger:           l,
	}
}

//easyjson:json
type CheckPaymentResponse struct {
	OrderID       string `json:"order_id"`
	PaymentID     string `json:"payment_id"`
	PaymentStatus string `json:"payment_status"`
}

// CancelOrder godoc
// @Summary 		Отмена заказа
// @Description		Пользователь отменяет свой заказ. Доступно только пока заказ не in_progress / not finished.
// @Tags			order
// @Accept			json
// @Produce			json
// @Param			id		path	string	true	"ID заказа"
// @Success			200		{object}  map[string]string			"Успешная отмена заказа"
// @Failure			400		{object}  response.ErrorResponse	"Отсутствует ID заказа или ошибка при отмене"
// @Failure			401		{object}  response.ErrorResponse	"Неавторизован"
// @Failure			403		{object}  response.ErrorResponse	"Доступ запрещен"
// @Failure			404		{object}  response.ErrorResponse	"Заказ не найден"
// @Failure			500		{object}  response.ErrorResponse	"Внутренняя ошибка сервера"
// @Router			/order/{id}/cancel [post]
func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	orderID := r.PathValue("id")
	if orderID == "" {
		response.Error(w, http.StatusBadRequest, "order id is required")
		return
	}

	if err := h.orderClient.CancelOrder(ctx, orderID, userID); err != nil {
		l.Error("cancel order failed", err)
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

// CheckPayment godoc
// @Summary 		Проверка статуса платежа
// @Description		Для dev-окружения: актуализирует статус платежа через YooKassa REST, эмулируя webhook.
// @Tags			order
// @Accept			json
// @Produce			json
// @Param			id		path	string	true	"ID заказа"
// @Success			200		{object}  CheckPaymentResponse		"Успешное обновление статуса"
// @Success			202		{object}  CheckPaymentResponse		"Платеж еще не готов (pending)"
// @Failure			400		{object}  response.ErrorResponse	"Отсутствует ID заказа"
// @Failure			401		{object}  response.ErrorResponse	"Неавторизован"
// @Failure			403		{object}  response.ErrorResponse	"Доступ запрещен"
// @Failure			404		{object}  response.ErrorResponse	"Заказ или платеж не найден"
// @Failure			500		{object}  response.ErrorResponse	"Внутренняя ошибка сервера"
// @Router			/order/{id}/check-payment [get]
func (h *OrderHandler) CheckPayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	orderID := r.PathValue("id")
	if orderID == "" {
		response.Error(w, http.StatusBadRequest, "order id is required")
		return
	}

	paymentID, err := h.orderClient.GetOrderPaymentID(ctx, orderID, userID)
	if err != nil {
		response.JSON(w, http.StatusAccepted, CheckPaymentResponse{
			OrderID:       orderID,
			PaymentStatus: "pending",
		})
		return
	}

	statusStr, err := h.paymentClient.RefreshPaymentStatus(ctx, paymentID)
	if err != nil {
		l.Error("refresh payment status failed", err)
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, CheckPaymentResponse{
		OrderID:       orderID,
		PaymentID:     paymentID,
		PaymentStatus: statusStr,
	})
}

// CreateOrder godoc
// @Summary      Создать заказ
// @Description  Создает заказ на основе корзины, рассчитывает долги (Splits) и запускает Сагу
// @Tags         order
// @Accept       json
// @Produce      json
// @Param        Idempotency-Key  header    string              true  "Ключ идемпотентности"
// @Param        input            body      CreateOrderRequest  true  "Данные для оформления"
// @Success      200    {object}  CreateOrderResponse "Заказ успешно создан (ссылки на оплату придут по WS)"
// @Failure      400    {object}  map[string]string   "Bad request, пустая корзина или есть ничейные блюда"
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
		PayForAll:          req.PayForAll,
		PayerMapping:       req.PayerMapping,
	}

	orderPublicID, err := h.orderClient.CreateOrder(ctx, userID, input, idemKey)
	if err != nil {
		if errors.Is(err, orderclient.ErrCartIsEmpty) {
			response.Error(w, http.StatusBadRequest, "Cart is empty")
			return
		} else if errors.Is(err, orderclient.ErrUnassignedItems) {
			response.Error(w, http.StatusBadRequest, "Cannot checkout: cart has unassigned items")
			return
		} else if errors.Is(err, orderclient.ErrAddressNotFound) {
			response.Error(w, http.StatusNotFound, "Address not found")
			return
		}

		l.Error("order creation failed via grpc", err)
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	l.Info("order created successfully", logger.String("order_id", orderPublicID))

	response.JSON(w, http.StatusOK, CreateOrderResponse{
		OrderID: orderPublicID,
	})
}

// GetMyOrders godoc
// @Summary      Получить историю заказов
// @Description  Возвращает историю заказов пользователя со всеми позициями и детализацией счетов (Splits)
// @Tags         order
// @Produce      json
// @Success      200    {array}   OrderHistoryResponse
// @Failure      401    {object}  map[string]string "Unauthorized"
// @Failure      500    {object}  map[string]string "Internal server error"
// @Security     ApiKeyAuth
// @Router       /api/profile/orders [get]
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

	dishIDSet := make(map[int64]struct{})
	for _, o := range orders {
		for _, item := range o.Items {
			dishIDSet[item.DishID] = struct{}{}
		}
	}

	dishMeta := make(map[int64]struct {
		Name     string
		ImageURL string
	}, len(dishIDSet))

	if len(dishIDSet) > 0 {
		dishIDs := make([]int64, 0, len(dishIDSet))
		for id := range dishIDSet {
			dishIDs = append(dishIDs, id)
		}

		dishes, derr := h.restaurantClient.GetDishesByIDs(ctx, dishIDs)
		if derr != nil {
			l.Error("failed to enrich order items with dish info", derr)
		} else {
			for _, d := range dishes {
				dishMeta[d.Id] = struct {
					Name     string
					ImageURL string
				}{Name: d.Name, ImageURL: d.ImageUrl}
			}
		}
	}

	resp := make([]OrderHistoryResponse, 0, len(orders))
	for _, o := range orders {

		items := make([]OrderDishDTO, 0, len(o.Items))
		for _, item := range o.Items {
			meta := dishMeta[item.DishID]
			items = append(items, OrderDishDTO{
				DishID:      item.DishID,
				Name:        meta.Name,
				ImageURL:    meta.ImageURL,
				Quantity:    item.Quantity,
				Price:       item.Price,
				OwnerUserID: item.OwnerUserID,
			})
		}

		splits := make([]OrderSplitDTO, 0, len(o.Splits))
		for _, split := range o.Splits {
			splits = append(splits, OrderSplitDTO{
				SplitID: split.SplitID,
				UserID:  split.UserID,
				Amount:  split.Amount,
				Status:  split.Status,
			})
		}

		resp = append(resp, OrderHistoryResponse{
			OrderID:            o.PublicID,
			RestaurantName:     o.RestaurantName,
			RestaurantImageURL: o.RestaurantLogoURL,
			TotalCost:          o.TotalCost,
			Status:             o.Status,
			CreatedAt:          o.CreatedAt.Format("02.01.2006"),
			Items:              items,
			Splits:             splits,
		})
	}

	response.JSON(w, http.StatusOK, resp)
}

// PayForFriend godoc
// @Summary      Оплатить за друга
// @Description  Перехватывает чужой счет (Сплит) и запускает по нему транзакцию оплаты
// @Tags         order
// @Accept       json
// @Produce      json
// @Param        id               path      string               true  "UUID сплита (счета)"
// @Param        Idempotency-Key  header    string               true  "Ключ идемпотентности"
// @Param        input            body      PayForFriendRequest  true  "Данные для оплаты"
// @Success      200    {object}  map[string]string    "Процесс оплаты запущен"
// @Failure      400    {object}  map[string]string    "Bad request"
// @Failure      401    {object}  map[string]string    "Unauthorized"
// @Failure      500    {object}  map[string]string    "Internal server error"
// @Security     ApiKeyAuth
// @Router       /api/orders/splits/{id}/pay [post]
func (h *OrderHandler) PayForFriend(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	splitID := r.PathValue("id")
	if splitID == "" {
		response.Error(w, http.StatusBadRequest, "split_id is required")
		return
	}

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		response.Error(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	var req PayForFriendRequest
	if err := request.JSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.orderClient.PayForFriend(ctx, splitID, userID, req.PaymentMethodID, idemKey)
	if err != nil {
		h.logger.Error("failed to pay for friend via grpc", err)
		response.Error(w, http.StatusInternalServerError, "failed to initiate payment")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "payment initiated"})
}

// TrackOrderWS godoc
// @Summary      WebSocket отслеживания заказа
// @Description  Подключение для получения live-обновлений по заказу и ссылок на оплату
// @Tags         order
// @Param        id   path  string  true  "ID заказа"
// @Success      101  "Switching Protocols"
// @Failure      400  {object}  map[string]string "Bad Request"
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Security     ApiKeyAuth
// @Router       /api/ws/orders/{id} [get]
func (h *OrderHandler) TrackOrderWS(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue("id")
	if orderID == "" {
		response.Error(w, http.StatusBadRequest, "order_id is required")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("failed to upgrade to websocket", err)
		return
	}

	h.wsManager.AddOrderConnection(orderID, conn)
}
