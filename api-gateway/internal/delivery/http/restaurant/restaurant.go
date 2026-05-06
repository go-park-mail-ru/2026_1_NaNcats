package restaurant

//go:generate easyjson $GOFILE

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/restaurantclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/response"
	pbRestaurant "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/restaurant"
)

// Захардкоженный список категорий
var hardcodedCategories = []CategoryResponse{
	{ID: "popular", Name: "Популярное", Emoji: "🔥"},
	{ID: "pizza", Name: "Пицца", Emoji: "🍕"},
	{ID: "sushi", Name: "Суши", Emoji: "🍣"},
	{ID: "burgers", Name: "Бургеры", Emoji: "🍔"},
	{ID: "desserts", Name: "Десерты", Emoji: "🍰"},
	{ID: "breakfast", Name: "Завтраки", Emoji: "🍳"},
	{ID: "health", Name: "Здоровье", Emoji: "🥦"},
	{ID: "coffee", Name: "Кофе", Emoji: "☕"},
	{ID: "steaks", Name: "Стейки", Emoji: "🥩"},
	{ID: "pasta", Name: "Паста", Emoji: "🍝"},
	{ID: "asian", Name: "Азиатская кухня", Emoji: "🥢"},
	{ID: "seafood", Name: "Морепродукты", Emoji: "🦞"},
	{ID: "fastfood", Name: "Фастфуд", Emoji: "🍟"},
	{ID: "russian", Name: "Русская кухня", Emoji: "🇷🇺"},
	{ID: "chinese", Name: "Китайская кухня", Emoji: "🥠"},
	{ID: "georgian", Name: "Грузинская кухня", Emoji: "🥙"},
	{ID: "home", Name: "Домашняя кухня", Emoji: "🏠"},
	{ID: "bread", Name: "Хлеб и выпечка", Emoji: "🥖"},
	{ID: "salads", Name: "Салаты", Emoji: "🥗"},
	{ID: "soups", Name: "Супы", Emoji: "🥣"},
}

var categoryNameMap = map[string]string{
	"popular":   "Популярное",
	"pizza":     "Пицца",
	"sushi":     "Суши",
	"burgers":   "Бургеры",
	"desserts":  "Десерты",
	"breakfast": "Завтраки",
	"health":    "Здоровье",
	"coffee":    "Кофе",
	"steaks":    "Стейки",
	"pasta":     "Паста",
	"asian":     "Азиатская кухня",
	"seafood":   "Морепродукты",
	"fastfood":  "Фастфуд",
	"russian":   "Русская",
	"chinese":   "Китайская",
	"georgian":  "Грузинская кухня",
	"home":      "Домашняя кухня",
	"bread":     "Хлеб и выпечка",
	"salads":    "Салаты",
	"soups":     "Супы",
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

// DishWithBrand расширяет DishResponse полем restaurant_brand_id -
// нужно для search-результатов, чтобы фронт мог отрисовать ссылку на ресторан.
//
//easyjson:json
type DishWithBrand struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	ImageURL          string `json:"image_url"`
	Price             int64  `json:"price"`
	RestaurantBrandID string `json:"restaurant_brand_id"`
}

// SearchAllResponse - общий ответ глобального поиска: рестораны и блюда вместе.
//
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

	searchQuery := query.Get("q")
	var (
		dishes []*pbRestaurant.Dish
	)
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

func (h *RestaurantHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, CategoriesResponse{Categories: hardcodedCategories})
}

func (h *RestaurantHandler) GetRestaurantBrandsListByCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	categorySlug := r.PathValue("slug")
	if categorySlug == "" {
		response.Error(w, http.StatusBadRequest, "Missing category slug")
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

	catName := categoryNameMap[categorySlug]
	if catName == "" {
		brands, err := h.restaurantClient.GetRestaurantBrandsList(ctx, int32(limit), int32(offset))
		if err != nil {
			l.Error("failed to get brands list", err)
			response.Error(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		dtoList := make([]RestaurantBrandResponse, 0, len(brands))
		for _, b := range brands {
			dtoList = append(dtoList, RestaurantBrandResponse{
				ID: strconv.FormatInt(b.Id, 10), Name: b.Name,
				Description: b.Description, PromotionTier: int(b.PromotionTier), LogoURL: b.LogoUrl,
			})
		}
		response.JSON(w, http.StatusOK, RestaurantBrandsResponse{RestaurantBrands: dtoList})
		return
	}

	brands, err := h.restaurantClient.GetRestaurantBrandsListByCategoryName(ctx, catName, int32(limit), int32(offset))
	if err != nil {
		l.Error("failed to get brands by category name", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	dtoList := make([]RestaurantBrandResponse, 0, len(brands))
	for _, b := range brands {
		dtoList = append(dtoList, RestaurantBrandResponse{
			ID: strconv.FormatInt(b.Id, 10), Name: b.Name,
			Description: b.Description, PromotionTier: int(b.PromotionTier), LogoURL: b.LogoUrl,
		})
	}
	response.JSON(w, http.StatusOK, RestaurantBrandsResponse{RestaurantBrands: dtoList})
}

func (h *RestaurantHandler) SearchRestaurants(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	q := r.URL.Query().Get("q")
	if q == "" {
		response.Error(w, http.StatusBadRequest, "Missing search query")
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
		brandList = append(brandList, RestaurantBrandResponse{
			ID: strconv.FormatInt(b.Id, 10), Name: b.Name,
			Description: b.Description, PromotionTier: int(b.PromotionTier), LogoURL: b.LogoUrl,
		})
	}

	dishList := make([]DishWithBrand, 0, len(dishes))
	for _, d := range dishes {
		dishList = append(dishList, DishWithBrand{
			ID:                strconv.FormatInt(d.Id, 10),
			Name:              d.Name,
			Description:       d.Description,
			ImageURL:          d.ImageUrl,
			Price:             d.Price,
			RestaurantBrandID: strconv.FormatInt(d.RestaurantBrandId, 10),
		})
	}

	response.JSON(w, http.StatusOK, SearchAllResponse{
		Restaurants: brandList,
		Dishes:      dishList,
	})
}
