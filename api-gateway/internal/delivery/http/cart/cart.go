package cart

//go:generate easyjson $GOFILE

import (
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/cartclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/request"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/response"
)

//easyjson:json
type CartItemDTO struct {
	DishID   int64  `json:"dish_id"`
	Name     string `json:"name,omitempty"`
	Price    int64  `json:"price,omitempty"`
	Quantity int32  `json:"quantity"`
	ImageURL string `json:"image_url,omitempty"`
}

//easyjson:json
type CartRequest struct {
	RestaurantID int64         `json:"restaurant_id"`
	Items        []CartItemDTO `json:"items"`
}

//easyjson:json
type CartResponse struct {
	RestaurantBrandID int64         `json:"restaurant_id"`
	Items             []CartItemDTO `json:"items"`
	TotalCost         int64         `json:"total_cost"`
}

type CartHandler struct {
	cartClient cartclient.CartClient
	logger     logger.Logger
}

func NewCartHandler(cc cartclient.CartClient, l logger.Logger) *CartHandler {
	return &CartHandler{
		cartClient: cc,
		logger:     l,
	}
}

// GetCart godoc
// @Summary      Получить корзину
// @Description  Возвращает текущую корзину авторизованного пользователя
// @Tags         cart
// @Produce      json
// @Success      200  {object}  CartResponse
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Security     ApiKeyAuth
// @Router       /api/cart [get]
func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	cart, err := h.cartClient.GetCart(ctx, userID)
	if err != nil {
		l.Error("failed to fetch cart via grpc", err)
		response.Error(w, http.StatusInternalServerError, "failed to get cart")
		return
	}

	cartResponse := CartResponse{
		RestaurantBrandID: cart.RestaurantID,
		TotalCost:         cart.TotalCost,
		Items:             make([]CartItemDTO, 0, len(cart.Items)),
	}

	for _, item := range cart.Items {
		cartResponse.Items = append(cartResponse.Items, CartItemDTO{
			DishID:   item.DishID,
			Name:     item.Name,
			Price:    item.Price,
			Quantity: item.Quantity,
			ImageURL: item.ImageURL,
		})
	}

	response.JSON(w, http.StatusOK, cartResponse)
}

// UpdateCart godoc
// @Summary      Обновить корзину
// @Description  Перезаписывает содержимое корзины пользователя новыми товарами
// @Tags         cart
// @Accept       json
// @Produce      json
// @Param        input  body      CartRequest  true  "Данные корзины"
// @Success      200    {object}  map[string]interface{} "Успешное обновление"
// @Failure      400    {object}  map[string]string "Неверный формат запроса или товары из разных ресторанов"
// @Failure      401    {object}  map[string]string "Unauthorized"
// @Failure      500    {object}  map[string]string "Internal server error"
// @Security     ApiKeyAuth
// @Router       /api/cart [put]
func (h *CartHandler) UpdateCart(w http.ResponseWriter, r *http.Request) {
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

	var reqCart CartRequest
	if err := request.JSON(r, &reqCart); err != nil {
		l.Warn("invalid update cart json", logger.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	clientItems := make([]cartclient.Item, 0, len(reqCart.Items))
	for _, it := range reqCart.Items {
		clientItems = append(clientItems, cartclient.Item{
			DishID:   it.DishID,
			Quantity: it.Quantity,
		})
	}

	err := h.cartClient.UpdateCart(ctx, userID, reqCart.RestaurantID, clientItems, idemKey)
	if err != nil {
		if errors.Is(err, cartclient.ErrInvalidCart) {
			l.Warn("cart update validation failed")
			response.Error(w, http.StatusBadRequest, "invalid cart data (check quantity or mixed restaurants)")
			return
		}
		l.Error("failed to sync cart via grpc", err)
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "cart updated successfully"})
}
