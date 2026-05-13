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

func parsePagination(r *http.Request) (int, int) {
	limit, offset := 20, 0
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
	return limit, offset
}

func toCategoryResponse(c restaurantclient.Category) CategoryResponse {
	return CategoryResponse{
		ID:    strconv.FormatInt(c.ID, 10),
		Name:  c.Name,
		Emoji: c.Emoji,
	}
}

func toRestaurantBrandResponse(b restaurantclient.RestaurantBrand) RestaurantBrandResponse {
	return RestaurantBrandResponse{
		ID:            strconv.FormatInt(b.ID, 10),
		Name:          b.Name,
		Description:   b.Description,
		PromotionTier: b.PromotionTier,
		LogoURL:       b.LogoURL,
	}
}

func toDishResponse(d restaurantclient.Dish) DishResponse {
	return DishResponse{
		ID:          strconv.FormatInt(d.ID, 10),
		Name:        d.Name,
		Description: d.Description,
		ImageURL:    d.ImageURL,
		Price:       d.Price,
	}
}

func toDishWithBrandResponse(d restaurantclient.Dish) DishWithBrand {
	return DishWithBrand{
		ID:                strconv.FormatInt(d.ID, 10),
		Name:              d.Name,
		Description:       d.Description,
		ImageURL:          d.ImageURL,
		Price:             d.Price,
		RestaurantBrandID: strconv.FormatInt(d.RestaurantBrandID, 10),
	}
}

//easyjson:json
type CategoryResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Emoji string `json:"emoji"`
}

//easyjson:json
type CategoriesResponse struct {
	Categories []CategoryResponse `json:"categories"`
}

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

//easyjson:json
type DishWithBrand struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	ImageURL          string `json:"image_url"`
	Price             int64  `json:"price"`
	RestaurantBrandID string `json:"restaurant_brand_id"`
}

//easyjson:json
type SearchAllResponse struct {
	Restaurants []RestaurantBrandResponse `json:"restaurants"`
	Dishes      []DishWithBrand           `json:"dishes"`
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

// GetRestaurantBrandsList godoc
// @Summary 		Получение списка ресторанов
// @Description		Возвращает список всех брендов ресторанов с поддержкой пагинации
// @Tags			restaurants
// @Produce			json
// @Param			limit	query	int		false	"Количество возвращаемых элементов (по умолчанию 20)"
// @Param			offset	query	int		false	"Смещение для пагинации (по умолчанию 0)"
// @Success			200		{object} RestaurantBrandsResponse	"Успешное получение списка ресторанов"
// @Failure			500		{object} response.ErrorResponse		"Внутренняя ошибка сервера"
// @Router			/restaurants [get]
func (h *RestaurantHandler) GetRestaurantBrandsList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	limit, offset := parsePagination(r)

	brands, err := h.restaurantClient.GetRestaurantBrandsList(ctx, int32(limit), int32(offset))
	if err != nil {
		l.Error("failed to get restaurant brand list", err)
		response.Error(w, http.StatusInternalServerError, "Get restaurant brand list error")
		return
	}

	dtoList := make([]RestaurantBrandResponse, 0, len(brands))
	for _, b := range brands {
		dtoList = append(dtoList, toRestaurantBrandResponse(b))
	}

	response.JSON(w, http.StatusOK, RestaurantBrandsResponse{RestaurantBrands: dtoList})
}

// GetRestaurantBrandByID godoc
// @Summary 		Получение ресторана по ID
// @Description		Возвращает детальную информацию о бренде ресторана по его уникальному идентификатору
// @Tags			restaurants
// @Produce			json
// @Param			id		path	int		true	"ID ресторана"
// @Success			200		{object} RestaurantBrandResponse	"Успешное получение ресторана"
// @Failure			400		{object} response.ErrorResponse		"Неверный формат ID"
// @Failure			404		{object} response.ErrorResponse		"Ресторан не найден"
// @Failure			500		{object} response.ErrorResponse		"Внутренняя ошибка сервера"
// @Router			/restaurants/{id} [get]
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

	response.JSON(w, http.StatusOK, toRestaurantBrandResponse(brand))
}

// GetDishesByRestaurantBrandID godoc
// @Summary 		Получение списка блюд ресторана
// @Description		Возвращает список блюд для конкретного ресторана с возможностью пагинации и поиска по названию
// @Tags			restaurants
// @Produce			json
// @Param			id		path	int		true	"ID ресторана"
// @Param			limit	query	int		false	"Количество возвращаемых элементов (по умолчанию 20)"
// @Param			offset	query	int		false	"Смещение для пагинации (по умолчанию 0)"
// @Param			q		query	string	false	"Поисковый запрос для фильтрации блюд"
// @Success			200		{object} DishesResponse				"Успешное получение списка блюд"
// @Failure			400		{object} response.ErrorResponse		"Неверный формат ID ресторана"
// @Failure			404		{object} response.ErrorResponse		"Блюда или ресторан не найдены"
// @Failure			500		{object} response.ErrorResponse		"Внутренняя ошибка сервера"
// @Router			/restaurants/{id}/dishes [get]
func (h *RestaurantHandler) GetDishesByRestaurantBrandID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	brandIDStr := r.PathValue("id")
	brandID, err := strconv.ParseInt(brandIDStr, 10, 64)
	if err != nil || brandID <= 0 {
		response.Error(w, http.StatusBadRequest, "Invalid restaurant brand id")
		return
	}

	limit, offset := parsePagination(r)
	searchQuery := r.URL.Query().Get("q")
	var dishes []restaurantclient.Dish

	if searchQuery != "" {
		dishes, err = h.restaurantClient.SearchDishesByBrand(ctx, brandID, searchQuery, int32(limit))
	} else {
		dishes, err = h.restaurantClient.GetDishesByRestaurantBrandID(ctx, brandID, int32(limit), int32(offset))
	}

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
		dtoList = append(dtoList, toDishResponse(d))
	}

	response.JSON(w, http.StatusOK, DishesResponse{Dishes: dtoList})
}

// GetCategories godoc
// @Summary 		Получение списка категорий
// @Description		Возвращает актуальный список доступных категорий кухни с их идентификаторами и эмодзи из БД
// @Tags			categories
// @Produce			json
// @Success			200		{object} CategoriesResponse			"Успешное получение списка категорий"
// @Failure			500		{object} response.ErrorResponse		"Внутренняя ошибка сервера"
// @Router			/categories [get]
func (h *RestaurantHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	cats, err := h.restaurantClient.GetCategories(ctx)
	if err != nil {
		l.Error("failed to get categories", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	dtoList := make([]CategoryResponse, 0, len(cats))
	for _, c := range cats {
		dtoList = append(dtoList, toCategoryResponse(c))
	}

	response.JSON(w, http.StatusOK, CategoriesResponse{Categories: dtoList})
}

// GetRestaurantBrandsListByCategory godoc
// @Summary 		Получение ресторанов по категории
// @Description		Возвращает список брендов ресторанов, отфильтрованный по переданному названию категории
// @Tags			categories
// @Produce			json
// @Param			slug	path	string	true	"Название (или слаг) категории"
// @Param			limit	query	int		false	"Количество возвращаемых элементов (по умолчанию 20)"
// @Param			offset	query	int		false	"Смещение для пагинации (по умолчанию 0)"
// @Success			200		{object} RestaurantBrandsResponse	"Успешное получение ресторанов по категории"
// @Failure			400		{object} response.ErrorResponse		"Отсутствует идентификатор категории"
// @Failure			500		{object} response.ErrorResponse		"Внутренняя ошибка сервера"
// @Router			/categories/{slug}/restaurants [get]
func (h *RestaurantHandler) GetRestaurantBrandsListByCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	categoryName := r.PathValue("slug")
	if categoryName == "" {
		response.Error(w, http.StatusBadRequest, "Missing category identifier")
		return
	}

	limit, offset := parsePagination(r)

	brands, err := h.restaurantClient.GetRestaurantBrandsByCategoryName(ctx, categoryName, int32(limit), int32(offset))
	if err != nil {
		l.Error("failed to get brands by category", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	dtoList := make([]RestaurantBrandResponse, 0, len(brands))
	for _, b := range brands {
		dtoList = append(dtoList, toRestaurantBrandResponse(b))
	}

	response.JSON(w, http.StatusOK, RestaurantBrandsResponse{RestaurantBrands: dtoList})
}

// SearchRestaurants godoc
// @Summary 		Глобальный поиск
// @Description		Осуществляет поиск по ресторанам и блюдам на основе переданного ключевого слова
// @Tags			search
// @Produce			json
// @Param			q		query	string	true	"Поисковый запрос"
// @Param			limit	query	int		false	"Количество возвращаемых элементов (по умолчанию 20)"
// @Param			offset	query	int		false	"Смещение для пагинации (по умолчанию 0, применяется к ресторанам)"
// @Success			200		{object} SearchAllResponse			"Успешный поиск (возвращает списки ресторанов и блюд)"
// @Failure			400		{object} response.ErrorResponse		"Отсутствует поисковый запрос"
// @Router			/search [get]
func (h *RestaurantHandler) SearchRestaurants(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	q := r.URL.Query().Get("q")
	if q == "" {
		response.Error(w, http.StatusBadRequest, "Missing search query")
		return
	}

	limit, offset := parsePagination(r)

	brands, err := h.restaurantClient.SearchRestaurantBrands(ctx, q, int32(limit), int32(offset))
	if err != nil {
		l.Error("failed to search restaurants", err)
		brands = nil
	}

	dishes, err := h.restaurantClient.SearchDishes(ctx, q, int32(limit))
	if err != nil {
		l.Error("failed to search dishes", err)
		dishes = nil
	}

	brandList := make([]RestaurantBrandResponse, 0, len(brands))
	for _, b := range brands {
		brandList = append(brandList, toRestaurantBrandResponse(b))
	}

	dishList := make([]DishWithBrand, 0, len(dishes))
	for _, d := range dishes {
		dishList = append(dishList, toDishWithBrandResponse(d))
	}

	response.JSON(w, http.StatusOK, SearchAllResponse{
		Restaurants: brandList,
		Dishes:      dishList,
	})
}
