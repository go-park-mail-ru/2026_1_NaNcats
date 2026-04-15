package restaurant

//go:generate easyjson $GOFILE

import (
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	restaurant "github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/restaurant"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/response"
)

//easyjson:json
type RestaurantBrandResponse struct {
	ID            string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name          string `json:"name" example:"KFC"`
	Description   string `json:"description" example:"Острые крылошки"`
	PromotionTier int    `json:"promotion_tier" example:"2"`
	LogoURL       string `json:"logo_url" example:"restaurants/logos/fjaun99f-8fna-h8ff-afvd-lmc01mca9jca.png"`
	BannerURL     string `json:"banner_url" example:"restaurangs/banners/fjaun99f-8fna-h8ff-afvd-lmc01mca9jca.png"`
}

//easyjson:json
type RestaurantBrandsResponse struct {
	RestaurantBrands []RestaurantBrandResponse `json:"restaurants"`
}

type restaurantBrandHandler struct {
	restaurantBrandUC restaurant.RestaurantBrandUseCase
	logger            domain.Logger
}

func NewRestaurantBrandHandler(rbuc restaurant.RestaurantBrandUseCase, logger domain.Logger) *restaurantBrandHandler {
	return &restaurantBrandHandler{
		restaurantBrandUC: rbuc,
		logger:            logger,
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
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	query := r.URL.Query()

	limit := 20
	offset := 0

	if qLimit := query.Get("limit"); qLimit != "" {
		if val, err := strconv.Atoi(qLimit); err == nil && val > 0 {
			limit = val
		} else {
			l.Debug("invalid limit query parameter, using default",
				domain.String("input", qLimit),
				domain.Int("default", limit),
			)
		}
	}

	if qOffset := query.Get("offset"); qOffset != "" {
		if val, err := strconv.Atoi(qOffset); err == nil && val >= 0 {
			offset = val
		} else {
			l.Debug("invalid offset query parameter, using default",
				domain.String("input", qOffset),
				domain.Int("default", offset),
			)
		}
	}

	restaurantBrandsList, err := h.restaurantBrandUC.GetRestaurantBrandsList(ctx, limit, offset)
	if err != nil {
		l.Error("failed to get restaurant brand list", err,
			domain.Int("limit", limit),
			domain.Int("offset", offset),
		)
		response.Error(w, http.StatusInternalServerError, "Get restaurant brand list error")
		return
	}

	dtoList := make([]RestaurantBrandResponse, 0, len(restaurantBrandsList))

	for _, currRestaurantBrand := range restaurantBrandsList {
		restResp := RestaurantBrandResponse{
			ID:            strconv.Itoa(currRestaurantBrand.ID),
			Name:          currRestaurantBrand.Name,
			Description:   currRestaurantBrand.Description,
			PromotionTier: currRestaurantBrand.PromotionTier,
			LogoURL:       currRestaurantBrand.LogoURL,
		}

		dtoList = append(dtoList, restResp)
	}

	l.Debug("successfully fetched restaurant brands",
		domain.Int("count", len(dtoList)),
		domain.Int("limit", limit),
		domain.Int("offset", offset),
	)

	resp := RestaurantBrandsResponse{
		RestaurantBrands: dtoList,
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *restaurantBrandHandler) GetRestaurantBrandByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		l.Warn("invalid restaurant id format", domain.String("id_str", idStr))
		response.Error(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	brand, err := h.restaurantBrandUC.GetRestaurantBrandByID(ctx, id)
	if err != nil {
		l.Warn("restaurant not found", domain.Int("id", id))
		response.Error(w, http.StatusNotFound, "Restaurant not found")
		return
	}

	l.Debug("successfully fetched restaurant brand", domain.Int("id", id))

	response.JSON(w, http.StatusOK, RestaurantBrandResponse{
		ID:          strconv.Itoa(brand.ID),
		Name:        brand.Name,
		Description: brand.Description,
		LogoURL:     brand.LogoURL,
	})
}
