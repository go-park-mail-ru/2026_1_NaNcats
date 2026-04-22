package restaurant

//go:generate easyjson $GOFILE

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/restaurantclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/response"
)

//easyjson:json
type RestaurantBrandResponse struct {
	ID            string `json:"id" example:"1"`
	Name          string `json:"name" example:"KFC"`
	Description   string `json:"description" example:"Острые крылышки"`
	PromotionTier int    `json:"promotion_tier" example:"2"`
	LogoURL       string `json:"logo_url"`
}

//easyjson:json
type RestaurantBrandsResponse struct {
	RestaurantBrands []RestaurantBrandResponse `json:"restaurants"`
}

//easyjson:json
type DishResponse struct {
	ID          string `json:"id" example:"1"`
	Name        string `json:"name" example:"Чизбургер"`
	Description string `json:"description" example:"Сочный бургер с сыром"`
	ImageURL    string `json:"image_url"`
	Price       int64  `json:"price" example:"199000000"`
}

//easyjson:json
type DishesResponse struct {
	Dishes []DishResponse `json:"dishes"`
}
type RestaurantHandler struct {
	restaurantClient restaurantclient.RestaurantClient
	logger           logger.Logger
}

func NewRestaurantHandler(rc restaurantclient.RestaurantClient, l logger.Logger) *RestaurantHandler {
	return &RestaurantHandler{
		restaurantClient: rc,
		logger:           l,
	}
}

func (h *RestaurantHandler) GetRestaurantBrandsList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	limit := 20
	offset := 0
	query := r.URL.Query()

	if qLimit := query.Get("limit"); qLimit != "" {
		if val, err := strconv.Atoi(qLimit); err == nil && val > 0 {
			limit = val
		}
	}

	if qOffset := query.Get("offset"); qOffset != "" {
		if val, err := strconv.Atoi(qOffset); err == nil && val >= 0 {
			offset = val
		}
	}

	brands, err := h.restaurantClient.GetRestaurantBrandsList(ctx, int32(limit), int32(offset))
	if err != nil {
		l.Error("failed to get restaurant brand list", err)
		response.Error(w, http.StatusInternalServerError, "Get restaurant brand list error")
		return
	}

	dtoList := make([]RestaurantBrandResponse, 0, len(brands))
	for _, b := range brands {
		dtoList = append(dtoList, RestaurantBrandResponse{
			ID:            strconv.FormatInt(b.Id, 10),
			Name:          b.Name,
			Description:   b.Description,
			PromotionTier: int(b.PromotionTier),
			LogoURL:       b.LogoUrl,
		})
	}

	response.JSON(w, http.StatusOK, RestaurantBrandsResponse{RestaurantBrands: dtoList})
}

func (h *RestaurantHandler) GetRestaurantBrandByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		response.Error(w, http.StatusBadRequest, "Invalid ID format")
		return
	}

	brand, err := h.restaurantClient.GetRestaurantBrandByID(ctx, id)
	if err != nil {
		if errors.Is(err, restaurantclient.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "Restaurant not found")
			return
		}
		l.Error("failed to get restaurant by id", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, RestaurantBrandResponse{
		ID:            strconv.FormatInt(brand.Id, 10),
		Name:          brand.Name,
		Description:   brand.Description,
		PromotionTier: int(brand.PromotionTier),
		LogoURL:       brand.LogoUrl,
	})
}

func (h *RestaurantHandler) GetDishesByRestaurantBrandID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	brandIDStr := r.PathValue("id")
	brandID, err := strconv.ParseInt(brandIDStr, 10, 64)
	if err != nil || brandID <= 0 {
		response.Error(w, http.StatusBadRequest, "Invalid restaurant brand id")
		return
	}

	limit := 20
	offset := 0
	query := r.URL.Query()

	if qLimit := query.Get("limit"); qLimit != "" {
		if val, err := strconv.Atoi(qLimit); err == nil && val > 0 {
			limit = val
		}
	}

	if qOffset := query.Get("offset"); qOffset != "" {
		if val, err := strconv.Atoi(qOffset); err == nil && val >= 0 {
			offset = val
		}
	}

	dishes, err := h.restaurantClient.GetDishesByRestaurantBrandID(ctx, brandID, int32(limit), int32(offset))
	if err != nil {
		if errors.Is(err, restaurantclient.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "Dishes or Restaurant not found")
			return
		}
		l.Error("failed to get dishes list", err)
		response.Error(w, http.StatusInternalServerError, "Get dishes list error")
		return
	}

	dtoList := make([]DishResponse, 0, len(dishes))
	for _, d := range dishes {
		dtoList = append(dtoList, DishResponse{
			ID:          strconv.FormatInt(d.Id, 10),
			Name:        d.Name,
			Description: d.Description,
			ImageURL:    d.ImageUrl,
			Price:       d.Price,
		})
	}

	response.JSON(w, http.StatusOK, DishesResponse{Dishes: dtoList})
}
