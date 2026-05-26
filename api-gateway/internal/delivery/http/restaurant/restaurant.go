package restaurant

//go:generate easyjson $GOFILE

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/restaurantclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/request"
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

//easyjson:json
type UpdateBrandRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
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

// GetRecommendations godoc
// @Summary 		Рекомендованные рестораны
// @Description		Подбор по эвристике «похожие категории» с fallback на trending за 7 дней. Для гостя — trending или топ по promotion_tier.
// @Tags			restaurants
// @Produce			json
// @Param			limit	query	int		false	"Сколько вернуть (по умолчанию 4)"
// @Success			200		{object} RestaurantBrandsResponse
// @Failure			500		{object} response.ErrorResponse
// @Router			/restaurants/recommendations [get]
func (h *RestaurantHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	limit := 4
	if q := r.URL.Query().Get("limit"); q != "" {
		if v, err := strconv.Atoi(q); err == nil && v > 0 && v <= 24 {
			limit = v
		}
	}

	var userID int64
	if id, ok := middleware.GetUserID(ctx); ok {
		userID = id
	}

	brands, err := h.restaurantClient.GetRecommendations(ctx, userID, int32(limit))
	if err != nil {
		l.Error("failed to get recommendations", err)
		response.Error(w, http.StatusInternalServerError, "Get recommendations error")
		return
	}

	out := make([]RestaurantBrandResponse, 0, len(brands))
	for _, b := range brands {
		out = append(out, toRestaurantBrandResponse(b))
	}
	response.JSON(w, http.StatusOK, RestaurantBrandsResponse{RestaurantBrands: out})
}

// GetRecommendedDishes godoc
// @Summary 		Рекомендованные блюда внутри ресторана
// @Description		Топ блюд бренда по продажам за 30 дней (paid|finished). Если данных нет — первые из меню.
// @Tags			restaurants
// @Produce			json
// @Param			id		path	int		true	"ID ресторана"
// @Param			limit	query	int		false	"Сколько вернуть (по умолчанию 4)"
// @Success			200		{object} DishesResponse
// @Failure			400		{object} response.ErrorResponse
// @Failure			500		{object} response.ErrorResponse
// @Router			/restaurants/brands/{id}/recommended-dishes [get]
func (h *RestaurantHandler) GetRecommendedDishes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	brandIDStr := r.PathValue("id")
	brandID, err := strconv.ParseInt(brandIDStr, 10, 64)
	if err != nil || brandID <= 0 {
		response.Error(w, http.StatusBadRequest, "Invalid restaurant id")
		return
	}

	limit := 4
	if q := r.URL.Query().Get("limit"); q != "" {
		if v, err := strconv.Atoi(q); err == nil && v > 0 && v <= 24 {
			limit = v
		}
	}

	var userID int64
	if id, ok := middleware.GetUserID(ctx); ok {
		userID = id
	}

	dishes, err := h.restaurantClient.GetRecommendedDishes(ctx, brandID, userID, int32(limit))
	if err != nil {
		l.Error("failed to get recommended dishes", err)
		response.Error(w, http.StatusInternalServerError, "Get recommended dishes error")
		return
	}

	out := make([]DishResponse, 0, len(dishes))
	for _, d := range dishes {
		out = append(out, toDishResponse(d))
	}
	response.JSON(w, http.StatusOK, DishesResponse{Dishes: out})
}

// UpdateBrand godoc
// @Summary 		Обновление текстового профиля ресторана (для владельцев)
// @Description		Позволяет владельцу обновить название и описание своего ресторанного бренда
// @Tags			owner, restaurants
// @Accept			json
// @Produce			json
// @Param			id				path	  int					true	"ID бренда"
// @Param			Idempotency-Key header    string				true	"Ключ идемпотентности"
// @Param			input			body	  UpdateBrandRequest	true	"Новые данные бренда"
// @Success			200				{object}  RestaurantBrandResponse
// @Failure			401				{object}  response.ErrorResponse "Неавторизован"
// @Failure			403				{object}  response.ErrorResponse "Отказ в доступе (не владелец этого ресторана)"
// @Failure			500				{object}  response.ErrorResponse "Внутренняя ошибка"
func (h *RestaurantHandler) UpdateBrand(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	// Читаем ID ресторана из параметров пути
	brandIDStr := r.PathValue("id")
	brandID, err := strconv.ParseInt(brandIDStr, 10, 64)
	if err != nil || brandID <= 0 {
		response.Error(w, http.StatusBadRequest, "Invalid restaurant brand id")
		return
	}

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		response.Error(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	var req UpdateBrandRequest
	if err := request.JSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Вызываем gRPC-клиент (внутренний UseCase в микросервисе проверит права)
	updatedBrand, err := h.restaurantClient.UpdateRestaurantBrand(ctx, brandID, req.Name, req.Description, nil, nil, idemKey)
	if err != nil {
		l.Error("failed to update restaurant brand via grpc", err)
		response.WriteError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, toRestaurantBrandResponse(updatedBrand))
}

// UpdateBrandLogo godoc
// @Summary 		Обновление логотипа ресторана (для владельцев)
// @Description		Загружает и заменяет логотип ресторанного бренда. Принимает multipart/form-data с полем 'logo'.
// @Tags			owner, restaurants
// @Accept			multipart/form-data
// @Produce			json
// @Param			id				path	  int		true	"ID бренда"
// @Param			Idempotency-Key header    string	true	"Ключ"
// @Param			logo			formData  file		true	"Файл логотипа (WEBP/PNG/JPG, до 5МБ)"
// @Success			200				{object}  RestaurantBrandResponse
func (h *RestaurantHandler) UpdateBrandLogo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	brandIDStr := r.PathValue("id")
	brandID, err := strconv.ParseInt(brandIDStr, 10, 64)
	if err != nil || brandID <= 0 {
		response.Error(w, http.StatusBadRequest, "Invalid restaurant brand id")
		return
	}

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		response.Error(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	// Парсим мультипарт форму (лимит 5 мб)
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		response.Error(w, http.StatusBadRequest, "file is too large (max 5MB)")
		return
	}

	file, _, err := r.FormFile("logo")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "failed to get 'logo' field from form")
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		l.Error("failed to read logo file", err)
		response.Error(w, http.StatusInternalServerError, "failed to process file")
		return
	}

	// Передаем байты изображения через gRPC
	updatedBrand, err := h.restaurantClient.UpdateRestaurantBrand(ctx, brandID, nil, nil, fileBytes, nil, idemKey)
	if err != nil {
		l.Error("failed to update brand logo via grpc", err)
		response.WriteError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, toRestaurantBrandResponse(updatedBrand))
}

// CreateDish godoc
// @Summary 		Добавление нового блюда в меню (для владельцев)
// @Description		Добавляет новое блюдо в указанный ресторанный бренд. Принимает multipart/form-data.
// @Tags			owner, dishes
// @Accept			multipart/form-data
// @Produce			json
// @Param			id				path	  int		true	"ID бренда (ресторана)"
// @Param			Idempotency-Key header    string	true	"Ключ"
// @Param			name			formData  string	true	"Название блюда"
// @Param			description		formData  string	false	"Описание блюда"
// @Param			price			formData  int		true	"Цена в сырых единицах (1 рубль = 1 000 000)"
// @Param			image			formData  file		false	"Изображение блюда (WEBP/PNG/JPG)"
// @Success			201				{object}  DishResponse
func (h *RestaurantHandler) CreateDish(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	brandIDStr := r.PathValue("id")
	brandID, err := strconv.ParseInt(brandIDStr, 10, 64)
	if err != nil || brandID <= 0 {
		response.Error(w, http.StatusBadRequest, "Invalid restaurant brand id")
		return
	}

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		response.Error(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	if err := r.ParseMultipartForm(5 << 20); err != nil {
		response.Error(w, http.StatusBadRequest, "form data is too large")
		return
	}

	name := r.FormValue("name")
	if name == "" {
		response.Error(w, http.StatusBadRequest, "dish name is required")
		return
	}

	description := r.FormValue("description")

	priceStr := r.FormValue("price")
	price, err := strconv.ParseInt(priceStr, 10, 64)
	if err != nil || price <= 0 {
		response.Error(w, http.StatusBadRequest, "invalid price value")
		return
	}

	// Изображение блюда является опциональным параметром
	var imageBytes []byte
	file, _, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		imageBytes, err = io.ReadAll(file)
		if err != nil {
			l.Error("failed to read dish image file", err)
			response.Error(w, http.StatusInternalServerError, "failed to process image file")
			return
		}
	}

	createdDish, err := h.restaurantClient.CreateDish(ctx, brandID, name, description, price, imageBytes, idemKey)
	if err != nil {
		l.Error("failed to create dish via grpc", err)
		response.WriteError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, toDishResponse(createdDish))
}

// CreateBrand godoc
// @Summary 		Регистрация нового ресторанного бренда (для владельцев)
// @Description		Позволяет пользователю с ролью 'owner' зарегистрировать новый ресторан. Принимает multipart/form-data.
// @Tags			owner, restaurants
// @Accept			multipart/form-data
// @Produce			json
// @Param			Idempotency-Key header    string	true	"Ключ идемпотентности"
// @Param			name			formData  string	true	"Название бренда"
// @Param			description		formData  string	false	"Описание бренда"
// @Param			logo			formData  file		false	"Файл логотипа бренда (WEBP/PNG/JPG)"
// @Success			201				{object}  RestaurantBrandResponse
func (h *RestaurantHandler) CreateBrand(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		response.Error(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	if err := r.ParseMultipartForm(5 << 20); err != nil {
		response.Error(w, http.StatusBadRequest, "form data is too large (max 5MB)")
		return
	}

	name := r.FormValue("name")
	if name == "" {
		response.Error(w, http.StatusBadRequest, "brand name is required")
		return
	}

	description := r.FormValue("description")

	var logoBytes []byte
	file, _, err := r.FormFile("logo")
	if err == nil {
		defer file.Close()
		logoBytes, err = io.ReadAll(file)
		if err != nil {
			l.Error("failed to read brand logo file", err)
			response.Error(w, http.StatusInternalServerError, "failed to process logo file")
			return
		}
	}

	// Передаем userID как овнера. Микросервис дополнительно проверит права
	// и принудительно перезапишет овнера из gRPC метаданных для безопасности.
	createdBrand, err := h.restaurantClient.CreateRestaurantBrand(ctx, userID, name, description, logoBytes, idemKey)
	if err != nil {
		l.Error("failed to create restaurant brand via grpc", err)
		response.WriteError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, toRestaurantBrandResponse(createdBrand))
}

// DeleteBrand godoc
// @Summary 		Удаление ресторанного бренда (для владельцев)
// @Description		Позволяет владельцу безвозвратно удалить свой ресторанный бренд и все связанные филиалы/меню
// @Tags			owner, restaurants
// @Param			id		path	  int	true	"ID бренда"
// @Success			200		{object}  response.MessageResponse "Ресторан успешно удален"
func (h *RestaurantHandler) DeleteBrand(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	brandIDStr := r.PathValue("id")
	brandID, err := strconv.ParseInt(brandIDStr, 10, 64)
	if err != nil || brandID <= 0 {
		response.Error(w, http.StatusBadRequest, "Invalid restaurant brand id")
		return
	}

	err = h.restaurantClient.DeleteRestaurantBrand(ctx, brandID)
	if err != nil {
		l.Error("failed to delete restaurant brand via grpc", err)
		response.WriteError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, response.MessageResponse{Message: "restaurant brand deleted successfully"})
}

// UpdateDish godoc
// @Summary 		Редактирование блюда в меню (для владельцев)
// @Description		Позволяет владельцу частично обновить поля блюда (название, описание, цену или изображение)
// @Tags			owner, dishes
// @Accept			multipart/form-data
// @Produce			json
// @Param			id				path	  int		true	"ID блюда"
// @Param			Idempotency-Key header    string	true	"Ключ"
// @Param			name			formData  string	false	"Новое название блюда"
// @Param			description		formData  string	false	"Новое описание блюда"
// @Param			price			formData  int		false	"Новая цена в сырых единицах (1 рубль = 1 000 000)"
// @Param			image			formData  file		false	"Новое изображение блюда (WEBP/PNG/JPG)"
// @Success			200				{object}  DishResponse
func (h *RestaurantHandler) UpdateDish(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	dishIDStr := r.PathValue("id")
	dishID, err := strconv.ParseInt(dishIDStr, 10, 64)
	if err != nil || dishID <= 0 {
		response.Error(w, http.StatusBadRequest, "Invalid dish id")
		return
	}

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		response.Error(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	if err := r.ParseMultipartForm(5 << 20); err != nil {
		response.Error(w, http.StatusBadRequest, "form data is too large")
		return
	}

	var namePtr *string
	if name := r.FormValue("name"); name != "" {
		namePtr = &name
	}

	var descPtr *string
	if desc := r.FormValue("description"); desc != "" {
		descPtr = &desc
	}

	var pricePtr *int64
	if priceStr := r.FormValue("price"); priceStr != "" {
		price, err := strconv.ParseInt(priceStr, 10, 64)
		if err != nil || price <= 0 {
			response.Error(w, http.StatusBadRequest, "invalid price value")
			return
		}
		pricePtr = &price
	}

	var imageBytes []byte
	file, _, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		imageBytes, err = io.ReadAll(file)
		if err != nil {
			l.Error("failed to read updated dish image file", err)
			response.Error(w, http.StatusInternalServerError, "failed to process image file")
			return
		}
	}

	updatedDish, err := h.restaurantClient.UpdateDish(ctx, dishID, namePtr, descPtr, pricePtr, imageBytes, idemKey)
	if err != nil {
		l.Error("failed to update dish via grpc", err)
		response.WriteError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, toDishResponse(updatedDish))
}

// DeleteDish godoc
// @Summary 		Удаление блюда из меню (для владельцев)
// @Description		Позволяет владельцу удалить блюдо из меню своего ресторана.
// @Tags			owner, dishes
// @Param			id		path	  int	true	"ID блюда"
// @Success			200		{object}  response.MessageResponse "Блюдо успешно удалено"
// @Failure			401		{object}  response.ErrorResponse "Неавторизован"
// @Failure			403		{object}  response.ErrorResponse "Отказ в доступе"
func (h *RestaurantHandler) DeleteDish(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	dishIDStr := r.PathValue("id")
	dishID, err := strconv.ParseInt(dishIDStr, 10, 64)
	if err != nil || dishID <= 0 {
		response.Error(w, http.StatusBadRequest, "Invalid dish id")
		return
	}

	err = h.restaurantClient.DeleteDish(ctx, dishID)
	if err != nil {
		l.Error("failed to delete dish via grpc", err)
		response.WriteError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, response.MessageResponse{Message: "dish deleted successfully"})
}
