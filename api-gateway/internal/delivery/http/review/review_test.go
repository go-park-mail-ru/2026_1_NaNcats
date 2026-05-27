package review

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/microcosm-cc/bluemonday"
	"github.com/stretchr/testify/assert"
)

func newHandler() *ReviewHandler {
	return NewReviewHandler(logger.NewNopLogger())
}

func TestNewReviewHandler_HasSeed(t *testing.T) {
	h := newHandler()
	assert.NotNil(t, h)
	// Должны быть два сид-отзыва для ресторана 1
	assert.Len(t, h.reviews[1], 2)
	assert.EqualValues(t, 3, h.nextID)
}

func TestReviewHandler_GetReviews(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectTotal    int
	}{
		{"Невалидный ID", "/restaurants/abc/reviews", http.StatusBadRequest, 0},
		{"ID 0", "/restaurants/0/reviews", http.StatusBadRequest, 0},
		{"Существующий ресторан с отзывами", "/restaurants/1/reviews", http.StatusOK, 2},
		{"Существующий ресторан без отзывов — пустой список", "/restaurants/99/reviews", http.StatusOK, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandler()
			mux := http.NewServeMux()
			mux.HandleFunc("GET /restaurants/{id}/reviews", h.GetReviews)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var resp ReviewsResponse
				assert.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
				assert.Equal(t, tt.expectTotal, resp.Total)
				assert.NotNil(t, resp.Reviews)
			}
		})
	}
}

func TestReviewHandler_CreateReview(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		body           any
		expectedStatus int
	}{
		{
			name:           "Невалидный ID — 400",
			path:           "/restaurants/abc/reviews",
			body:           CreateReviewRequest{Rating: 5, Comment: "ok"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Битый JSON — 400",
			path:           "/restaurants/1/reviews",
			body:           "definitely not a struct",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Рейтинг 0 — 400",
			path:           "/restaurants/1/reviews",
			body:           CreateReviewRequest{Rating: 0, Comment: "хорошо"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Рейтинг 6 — 400",
			path:           "/restaurants/1/reviews",
			body:           CreateReviewRequest{Rating: 6, Comment: "хорошо"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Пустой Comment — 400",
			path:           "/restaurants/1/reviews",
			body:           CreateReviewRequest{Rating: 5, Comment: ""},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Без AuthorName — успех с подстановкой 'Аноним'",
			path:           "/restaurants/1/reviews",
			body:           CreateReviewRequest{Rating: 4, Comment: "норм"},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Полный успех",
			path:           "/restaurants/1/reviews",
			body:           CreateReviewRequest{AuthorName: "Андрей", Rating: 5, Comment: "топчик"},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandler()
			mux := http.NewServeMux()
			mux.HandleFunc("POST /restaurants/{id}/reviews", h.CreateReview)

			raw, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, tt.path, bytes.NewBuffer(raw))
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var rev Review
				assert.NoError(t, json.NewDecoder(w.Body).Decode(&rev))
				assert.NotZero(t, rev.ID)
				assert.NotEmpty(t, rev.CreatedAt)
				assert.NotEmpty(t, rev.AuthorName, "пустое имя должно подменяться на 'Аноним'")
			}
		})
	}
}

func TestCreateReviewRequest_Sanitize(t *testing.T) {
	req := &CreateReviewRequest{
		AuthorName: "<script>alert(1)</script>Андрей",
		Comment:    "<b>жирный</b> <script>bad</script>",
	}
	p := bluemonday.UGCPolicy()
	req.Sanitize(p)
	assert.NotContains(t, req.AuthorName, "<script>")
	assert.NotContains(t, req.Comment, "<script>")
	assert.Contains(t, req.AuthorName, "Андрей")
	assert.Contains(t, req.Comment, "жирный")
}
