package handler

import (
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/response"
)

type RestaurantBrandResponse struct {
	ID            string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name          string `json:"name" example:"KFC"`
	Description   string `json:"description" example:"Острые крылошки"`
	PromotionTier int    `json:"promotion_tier" example:"2"`
}

type RestaurantBrandsResponse struct {
	RestaurantBrands []RestaurantBrandResponse `json:"restaurants"`
}

type restaurantBrandHandler struct {
	restaurantBrandUC usecase.RestaurantBrandUseCase
}

func NewRestaurantBrandHandler(rbuc usecase.RestaurantBrandUseCase) *restaurantBrandHandler {
	return &restaurantBrandHandler{
		restaurantBrandUC: rbuc,
	}
}

func (h *restaurantBrandHandler) GetRestaurantBrandsList(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	// Значения по дефолту
	limit := 20
	offset := 0

	// Считываем параметры
	// читаем limit
	if qLimit := query.Get("limit"); qLimit != "" {
		if val, err := strconv.Atoi(qLimit); err == nil {
			limit = val
		}
	}

	// читаем offset
	if qOffset := query.Get("offset"); qOffset != "" {
		if val, err := strconv.Atoi(qOffset); err == nil {
			offset = val
		}
	}

	ctx := r.Context()
	restaurantBrandsList := h.restaurantBrandUC.GetRestaurantBrandsList(ctx, limit, offset)

	dtoList := make([]RestaurantBrandResponse, 0, len(restaurantBrandsList))

	for _, currRestaurantBrand := range restaurantBrandsList {
		restResp := RestaurantBrandResponse{
			ID:            currRestaurantBrand.ID.URN(),
			Name:          currRestaurantBrand.Name,
			Description:   currRestaurantBrand.Description,
			PromotionTier: currRestaurantBrand.PromotionTier,
		}

		dtoList = append(dtoList, restResp)
	}

	resp := RestaurantBrandsResponse{
		RestaurantBrands: dtoList,
	}

	response.JSON(w, http.StatusOK, resp)
}
