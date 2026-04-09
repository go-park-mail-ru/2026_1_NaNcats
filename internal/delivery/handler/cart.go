package handler

import (
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/request"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/response"
)

type CartItemDTO struct {
	DishID   int    `json:"dish_id"`
	Name     string `json:"name,omitempty"`
	Price    int64  `json:"price,omitempty"`
	Quantity int    `json:"quantity"`
	ImageURL string `json:"image_url,omitempty"`
}

type CartRequest struct {
	RestaurantID int           `json:"restaurant_id"`
	Items        []CartItemDTO `json:"items"`
}

type CartResponse struct {
	Items             []CartItemDTO `json:"items"`
	RestaurantBrandID int           `json:"restaurant_id"`
	TotalCost         int64         `json:"total_cost"`
	UpdatedAt         string        `json:"updated_at"`
}

type cartHandler struct {
	cartUC usecase.CartUseCase
	logger domain.Logger
}

func NewCartHandler(cuc usecase.CartUseCase, l domain.Logger) *cartHandler {
	return &cartHandler{
		cartUC: cuc,
		logger: l,
	}
}

func (h *cartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		l.Error("auth middleware missed userID in route", err, nil)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	cart, total_cost, err := h.cartUC.GetCart(ctx, userID)
	if err != nil {
		l.Error("get cart failed", err, nil)
	}

	cartResponse := CartResponse{
		Items:             make([]CartItemDTO, 0, len(cart.Items)),
		RestaurantBrandID: cart.RestaurantBrandID,
		TotalCost:         total_cost,
		UpdatedAt:         cart.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	for _, item := range cart.Items {
		cartResponse.Items = append(cartResponse.Items, CartItemDTO{
			DishID:   item.DishID,
			Name:     item.Name,
			Price:    item.Price,
			Quantity: item.Quantity,
			ImageURL: "/api/images/" + item.ImageURL, // здесь изменить, добавить маппер
		})
	}

	l.Info("get cart success", map[string]any{
		"user_id":       userID,
		"items_count":   len(cart.Items),
		"total_cost":    total_cost,
		"restaurant_id": cart.RestaurantBrandID,
	})

	response.JSON(w, http.StatusOK, cartResponse)
}

func (h *cartHandler) UpdateCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		l.Error("auth middleware missed userID in route", err, nil)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	var reqCart CartRequest
	err = request.JSON(r, &reqCart, l)
	if err != nil {
		l.Info("invalid update cart json", map[string]any{"error": err.Error()})
		response.Error(w, http.StatusBadRequest, "internal server error")
		return
	}

	domainCart := domain.Cart{
		UserID:            userID,
		RestaurantBrandID: reqCart.RestaurantID,
		Items:             make([]domain.CartItem, 0, len(reqCart.Items)),
	}

	for _, it := range reqCart.Items {
		domainCart.Items = append(domainCart.Items, domain.CartItem{
			DishID:   it.DishID,
			Quantity: it.Quantity,
		})
	}

	err = h.cartUC.UpdateCart(ctx, userID, domainCart)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidQuantity):
			response.Error(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, domain.ErrMultipleRestaurants):
			response.Error(w, http.StatusBadRequest, err.Error())
		default:
			l.Error("failed to sync cart", err, map[string]any{"user_id": userID})
			response.Error(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	l.Info("cart synced successfully", map[string]any{
		"user_id":       userID,
		"items_count":   len(reqCart.Items),
		"restaurant_id": reqCart.RestaurantID,
	})

	response.JSON(w, http.StatusOK, nil)
}
