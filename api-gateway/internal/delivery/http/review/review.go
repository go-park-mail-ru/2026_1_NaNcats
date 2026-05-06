package review

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/request"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/response"
	"github.com/microcosm-cc/bluemonday"
)

//easyjson:json
type Review struct {
	ID           int64  `json:"id" example:"1"`
	RestaurantID int64  `json:"restaurant_id" example:"1"`
	AuthorName   string `json:"author_name" example:"Алексей"`
	Rating       int    `json:"rating" example:"5"`
	Comment      string `json:"comment" example:"Очень вкусно, доставили быстро!"`
	CreatedAt    string `json:"created_at" example:"2026-04-20"`
}

//easyjson:json
type CreateReviewRequest struct {
	AuthorName string `json:"author_name" example:"Алексей"`
	Rating     int    `json:"rating" example:"5"`
	Comment    string `json:"comment" example:"Очень вкусно, доставили быстро!"`
}

func (req *CreateReviewRequest) Sanitize(p *bluemonday.Policy) {
	req.AuthorName = p.Sanitize(req.AuthorName)
	req.Comment = p.Sanitize(req.Comment)
}

//easyjson:json
type ReviewsResponse struct {
	Reviews []Review `json:"reviews"`
	Total   int      `json:"total" example:"2"`
}

type ReviewHandler struct {
	mu      sync.RWMutex
	reviews map[int64][]Review
	nextID  int64
	logger  logger.Logger
}

func NewReviewHandler(l logger.Logger) *ReviewHandler {
	h := &ReviewHandler{
		reviews: make(map[int64][]Review),
		nextID:  1,
		logger:  l,
	}
	h.reviews[1] = []Review{
		{ID: 1, RestaurantID: 1, AuthorName: "Алексей", Rating: 5, Comment: "Очень вкусно, доставили быстро!", CreatedAt: "2026-04-20"},
		{ID: 2, RestaurantID: 1, AuthorName: "Мария", Rating: 4, Comment: "Хороший ресторан, рекомендую.", CreatedAt: "2026-04-18"},
	}
	h.nextID = 3
	return h
}

// GetReviews godoc
// @Summary 		Получение списка отзывов
// @Description		Возвращает список всех отзывов для конкретного ресторана по его ID
// @Tags			review
// @Produce			json
// @Param			id		path	int		true	"ID ресторана"
// @Success			200		{object} ReviewsResponse			"Успешное получение отзывов"
// @Failure			400		{object} response.ErrorResponse		"Неверный ID ресторана"
// @Router			/restaurants/{id}/reviews [get]
func (h *ReviewHandler) GetReviews(w http.ResponseWriter, r *http.Request) {
	restaurantIDStr := r.PathValue("id")
	restaurantID, err := strconv.ParseInt(restaurantIDStr, 10, 64)
	if err != nil || restaurantID <= 0 {
		response.Error(w, http.StatusBadRequest, "Invalid restaurant id")
		return
	}

	h.mu.RLock()
	reviews := h.reviews[restaurantID]
	h.mu.RUnlock()

	if reviews == nil {
		reviews = []Review{}
	}

	response.JSON(w, http.StatusOK, ReviewsResponse{Reviews: reviews, Total: len(reviews)})
}

// CreateReview godoc
// @Summary 		Создание отзыва
// @Description		Создает новый отзыв для указанного ресторана с оценкой и комментарием
// @Tags			review
// @Accept			json
// @Produce			json
// @Param			id		path	int						true	"ID ресторана"
// @Param			input	body	CreateReviewRequest		true	"Данные для создания отзыва"
// @Success			201		{object} Review						"Отзыв успешно создан"
// @Failure			400		{object} response.ErrorResponse		"Неверный формат JSON, ID ресторана или ошибка валидации полей"
// @Router			/restaurants/{id}/reviews [post]
func (h *ReviewHandler) CreateReview(w http.ResponseWriter, r *http.Request) {
	restaurantIDStr := r.PathValue("id")
	restaurantID, err := strconv.ParseInt(restaurantIDStr, 10, 64)
	if err != nil || restaurantID <= 0 {
		response.Error(w, http.StatusBadRequest, "Invalid restaurant id")
		return
	}

	var req CreateReviewRequest
	if err := request.JSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if req.Rating < 1 || req.Rating > 5 {
		response.Error(w, http.StatusBadRequest, "Rating must be between 1 and 5")
		return
	}
	if req.Comment == "" {
		response.Error(w, http.StatusBadRequest, "Comment is required")
		return
	}
	if req.AuthorName == "" {
		req.AuthorName = "Аноним"
	}

	review := Review{
		RestaurantID: restaurantID,
		AuthorName:   req.AuthorName,
		Rating:       req.Rating,
		Comment:      req.Comment,
		CreatedAt:    time.Now().Format("2006-01-02"),
	}

	h.mu.Lock()
	review.ID = h.nextID
	h.nextID++
	h.reviews[restaurantID] = append(h.reviews[restaurantID], review)
	h.mu.Unlock()

	response.JSON(w, http.StatusCreated, review)
}
