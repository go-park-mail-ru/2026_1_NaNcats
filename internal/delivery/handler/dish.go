package handler

//go:generate easyjson $GOFILE

import (
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/response"
)

//easyjson:json
type DishResponse struct {
	ID          string `json:"id" example:"1"`
	Name        string `json:"name" example:"Чизбургер"`
	Description string `json:"description" example:"Сочный бургер с сыром"`
	ImageURL    string `json:"image_url" example:"/api/images/dishes/cheeseburger.png"`
	Price       int64  `json:"price" example:"19900"`
}

//easyjson:json
type DishesResponse struct {
	Dishes []DishResponse `json:"dishes"`
}

type dishHandler struct {
	dishUC usecase.DishUseCase
	logger domain.Logger
}

func NewDishHandler(duc usecase.DishUseCase, logger domain.Logger) *dishHandler {
	return &dishHandler{
		dishUC: duc,
		logger: logger,
	}
}

// GetDishesByRestaurantBrandID godoc
// @Summary 		Получение списка блюд ресторана
// @Description		Возвращает список блюд ресторана (по restaurant_brand_id) с поддержкой пагинации (limit и offset)
// @Tags				restaurants
// @Produce				json
// @Param				id		path	  int	true	"ID бренда ресторана (restaurant_brand_id)"
// @Param				limit	query	  int	false	"Количество получаемых блюд"	default(20)
// @Param				offset	query	  int	false	"Смещение от начала списка"     	default(0)
// @Success				200		{object}  DishesResponse			"Успешное получение списка блюд"
// @Router				/restaurants/brands/{id}/dishes [get]
func (h *dishHandler) GetDishesByRestaurantBrandID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	query := r.URL.Query()

	// Значения по дефолту
	limit := 20
	offset := 0

	// limit
	if qLimit := query.Get("limit"); qLimit != "" {
		if val, err := strconv.Atoi(qLimit); err == nil && val > 0 {
			limit = val
		} else {
			l.Debug("invalid limit query parameter, using default", map[string]any{
				"input":   qLimit,
				"default": limit,
			})
		}
	}

	// offset
	if qOffset := query.Get("offset"); qOffset != "" {
		if val, err := strconv.Atoi(qOffset); err == nil && val >= 0 {
			offset = val
		} else {
			l.Debug("invalid offset query parameter, using default", map[string]any{
				"input":   qOffset,
				"default": offset,
			})
		}
	}

	// Получение id из пути
	restaurantBrandIDStr := r.PathValue("id")

	restaurantBrandID, err := strconv.Atoi(restaurantBrandIDStr)
	if err != nil || restaurantBrandID <= 0 {
		l.Debug("invalid restaurant brand id path parameter", map[string]any{
			"input": restaurantBrandIDStr,
		})
		response.Error(w, http.StatusBadRequest, "Invalid restaurant brand id")
		return
	}

	dishes, err := h.dishUC.GetDishesByRestaurantBrandID(ctx, restaurantBrandID, limit, offset)
	if err != nil {
		l.Error("Failed to get dishes list", err, map[string]any{
			"restaurant_brand_id": restaurantBrandID,
			"limit":               limit,
			"offset":              offset,
		})
		response.Error(w, http.StatusInternalServerError, "Get dishes list error")
		return
	}

	dto := make([]DishResponse, 0, len(dishes))
	for _, d := range dishes {
		img := d.ImageURL
		if img == "" {
			img = "/api/images/default/dish.png"
		} else {
			img = "/api/images/" + img
		}

		dto = append(dto, DishResponse{
			ID:          strconv.Itoa(d.ID),
			Name:        d.Name,
			Description: d.Description,
			ImageURL:    img,
			Price:       d.Price,
		})
	}

	l.Debug("successfully fetched dishes", map[string]any{
		"count":               len(dto),
		"restaurant_brand_id": restaurantBrandID,
		"limit":               limit,
		"offset":              offset,
	})

	response.JSON(w, http.StatusOK, DishesResponse{Dishes: dto})
}
