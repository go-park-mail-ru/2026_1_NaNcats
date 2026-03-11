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
	LogoURL       string `json:"logo_url" example:"restaurants/logos/fjaun99f-8fna-h8ff-afvd-lmc01mca9jca.png"`
	BannerURL     string `json:"banner_url" example:"restaurangs/banners/fjaun99f-8fna-h8ff-afvd-lmc01mca9jca.png"`
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

// GetRestaurantBrandsList godoc
// @Summary 		Получение списка ресторанов
// @Description		Возвращает список брендов ресторанов с поддержкой пагинации (limit и offset)
// @Tags				restaurants
// @Produce				json
// @Param				limit	query	  int	false	"Количество получаемых ресторанов"	default(20)
// @Param				offset	query	  int	false	"Смещение от начала списка"     	default(0)
// @Success				200		{object}  RestaurantBrandsResponse			"Успешное получение списка ресторанов"
// @Router				/restaurants/brands [get]
func (h *restaurantBrandHandler) GetRestaurantBrandsList(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	// Значения по дефолту
	limit := 20
	offset := 0

	// Считываем параметры
	// читаем limit
	if qLimit := query.Get("limit"); qLimit != "" {
		if val, err := strconv.Atoi(qLimit); err == nil && val > 0 {
			limit = val
		}
	}

	// читаем offset
	if qOffset := query.Get("offset"); qOffset != "" {
		if val, err := strconv.Atoi(qOffset); err == nil && val > 0 {
			offset = val
		}
	}

	ctx := r.Context()
	restaurantBrandsList := h.restaurantBrandUC.GetRestaurantBrandsList(ctx, limit, offset)

	dtoList := make([]RestaurantBrandResponse, 0, len(restaurantBrandsList))

	for _, currRestaurantBrand := range restaurantBrandsList {
		if currRestaurantBrand.LogoURL == "" {
			currRestaurantBrand.LogoURL = "/api/images/default/logo.png"
		} else {
			currRestaurantBrand.LogoURL = "/api/images/" + currRestaurantBrand.LogoURL
		}

		if currRestaurantBrand.BannerURL == "" {
			currRestaurantBrand.BannerURL = "/api/images/default/banner.png"
		} else {
			currRestaurantBrand.BannerURL = "/api/images/" + currRestaurantBrand.BannerURL
		}

		restResp := RestaurantBrandResponse{
			ID:            currRestaurantBrand.ID.String(),
			Name:          currRestaurantBrand.Name,
			Description:   currRestaurantBrand.Description,
			PromotionTier: currRestaurantBrand.PromotionTier,
			LogoURL:       currRestaurantBrand.LogoURL,
			BannerURL:     currRestaurantBrand.BannerURL,
		}

		dtoList = append(dtoList, restResp)
	}

	resp := RestaurantBrandsResponse{
		RestaurantBrands: dtoList,
	}

	response.JSON(w, http.StatusOK, resp)
}
