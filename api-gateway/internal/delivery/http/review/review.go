package review

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/response"
)

type Review struct {
	ID           int64  `json:"id"`
	RestaurantID int64  `json:"restaurant_id"`
	AuthorName   string `json:"author_name"`
	Rating       int    `json:"rating"`
	Comment      string `json:"comment"`
	CreatedAt    string `json:"created_at"`
}

type CreateReviewRequest struct {
	AuthorName string `json:"author_name"`
	Rating     int    `json:"rating"`
	Comment    string `json:"comment"`
}

type ReviewsResponse struct {
	Reviews []Review `json:"reviews"`
	Total   int      `json:"total"`
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
	// Seed some demo reviews
	h.reviews[1] = []Review{
		{ID: 1, RestaurantID: 1, AuthorName: "Алексей", Rating: 5, Comment: "Очень вкусно, доставили быстро!", CreatedAt: "2026-04-20"},
		{ID: 2, RestaurantID: 1, AuthorName: "Мария", Rating: 4, Comment: "Хороший ресторан, рекомендую.", CreatedAt: "2026-04-18"},
	}
	h.nextID = 3
	return h
}

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

func (h *ReviewHandler) CreateReview(w http.ResponseWriter, r *http.Request) {
	restaurantIDStr := r.PathValue("id")
	restaurantID, err := strconv.ParseInt(restaurantIDStr, 10, 64)
	if err != nil || restaurantID <= 0 {
		response.Error(w, http.StatusBadRequest, "Invalid restaurant id")
		return
	}

	var req CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
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
