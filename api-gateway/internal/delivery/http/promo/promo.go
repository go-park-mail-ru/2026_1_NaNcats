package promo

//go:generate easyjson $GOFILE

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/request"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/response"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrPromoNotFound  = errors.New("promo not found")
	ErrAlreadyBound   = errors.New("already bound")
	ErrPromoExpired   = errors.New("promo has expired")
	ErrMaxUsesReached = errors.New("max uses reached")
	ErrMinOrderAmount = errors.New("order amount is below minimum")
	ErrWrongBrand     = errors.New("promo is not valid for this restaurant")
)

//easyjson:json
type PromoCodeDTO struct {
	ID                 int64   `json:"id"`
	Code               string  `json:"code"`
	Title              string  `json:"title"`
	DiscountPercent    *int32  `json:"discount_percent"`
	DiscountAmount     *int64  `json:"discount_amount"`
	MinOrderAmount     int64   `json:"min_order_amount"`
	ExpiresAt          *string `json:"expires_at"`
	RestaurantBrandIDs []int64 `json:"restaurant_brand_ids"`
}

//easyjson:json
type PromoCodeDTOList []PromoCodeDTO

//easyjson:json
type BindPromoRequest struct {
	Code string `json:"code"`
}

//easyjson:json
type ValidatePromoRequest struct {
	Code              string `json:"code"`
	RestaurantBrandID int64  `json:"restaurant_brand_id"`
	OrderAmount       int64  `json:"order_amount"`
	DeliveryCost      int64  `json:"delivery_cost"`
	ServiceFee        int64  `json:"service_fee"`
}

//easyjson:json
type ValidatePromoResponse struct {
	Valid    bool   `json:"valid"`
	Discount int64  `json:"discount"`
	Reason   string `json:"reason"`
}

//easyjson:json
type UsePromoRequest struct {
	Code string `json:"code"`
	// OrderPublicID — публичный id заказа (тот, что отдаётся фронту).
	// Внутренний order.id резолвится на стороне хендлера.
	OrderPublicID string `json:"order_public_id"`
}

type PromoHandler struct {
	pool   *pgxpool.Pool
	logger logger.Logger
}

func NewPromoHandler(pool *pgxpool.Pool, l logger.Logger) *PromoHandler {
	return &PromoHandler{pool: pool, logger: l}
}

// GetUserPromos отдаёт промокоды, привязанные к пользователю.
func (h *PromoHandler) GetUserPromos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	promos, err := h.getUserPromos(ctx, userID)
	if err != nil {
		l.Error("failed to get user promos", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, PromoCodeDTOList(promos))
}

// BindPromo привязывает промокод к пользователю по коду.
func (h *PromoHandler) BindPromo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req BindPromoRequest
	if err := request.JSON(r, &req); err != nil {
		l.Warn("failed to decode bind promo request", logger.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Code == "" {
		response.Error(w, http.StatusBadRequest, "code is required")
		return
	}

	promo, err := h.bindPromo(ctx, userID, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, ErrPromoNotFound):
			response.Error(w, http.StatusBadRequest, "promo not found")
		case errors.Is(err, ErrAlreadyBound):
			response.Error(w, http.StatusConflict, "already bound")
		case errors.Is(err, ErrPromoExpired):
			response.Error(w, http.StatusBadRequest, "promo has expired")
		default:
			l.Error("failed to bind promo", err)
			response.Error(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	response.JSON(w, http.StatusOK, promo)
}

// ValidatePromo проверяет промокод для конкретного заказа и считает скидку.
func (h *PromoHandler) ValidatePromo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req ValidatePromoRequest
	if err := request.JSON(r, &req); err != nil {
		l.Warn("failed to decode validate promo request", logger.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Code == "" {
		response.Error(w, http.StatusBadRequest, "code is required")
		return
	}

	result, err := h.validatePromo(ctx, userID, req.Code, req.RestaurantBrandID, req.OrderAmount, req.DeliveryCost, req.ServiceFee)
	if err != nil {
		l.Error("failed to validate promo", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, result)
}

// UsePromo фиксирует использование промокода после оформления заказа.
func (h *PromoHandler) UsePromo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req UsePromoRequest
	if err := request.JSON(r, &req); err != nil {
		l.Warn("failed to decode use promo request", logger.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Code == "" {
		response.Error(w, http.StatusBadRequest, "code is required")
		return
	}

	err := h.recordPromoUsage(ctx, userID, req.Code, req.OrderPublicID)
	if err != nil {
		switch {
		case errors.Is(err, ErrPromoNotFound):
			response.Error(w, http.StatusBadRequest, "promo not found")
		case errors.Is(err, ErrMaxUsesReached):
			response.Error(w, http.StatusConflict, "max uses reached")
		default:
			l.Error("failed to record promo usage", err)
			response.Error(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GetRestaurantPromos отдаёт действующие промокоды ресторана.
// Публичный эндпоинт: баннер промокода виден всем, в том числе гостям.
func (h *PromoHandler) GetRestaurantPromos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	brandID, err := strconv.ParseInt(r.URL.Query().Get("brand_id"), 10, 64)
	if err != nil || brandID <= 0 {
		response.Error(w, http.StatusBadRequest, "brand_id is required")
		return
	}

	promos, err := h.getRestaurantPromos(ctx, brandID)
	if err != nil {
		l.Error("failed to get restaurant promos", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, PromoCodeDTOList(promos))
}

// --- repository helpers ---

// scanPromoList вычитывает строки промокодов (в фиксированном порядке колонок)
// и догружает к каждому список ресторанных брендов.
func (h *PromoHandler) scanPromoList(ctx context.Context, rows pgx.Rows) ([]PromoCodeDTO, error) {
	defer rows.Close()

	var promos []PromoCodeDTO
	for rows.Next() {
		var dto PromoCodeDTO
		var expiresAt *time.Time
		if err := rows.Scan(&dto.ID, &dto.Code, &dto.Title, &dto.DiscountPercent,
			&dto.DiscountAmount, &dto.MinOrderAmount, &expiresAt); err != nil {
			return nil, err
		}
		if expiresAt != nil {
			s := expiresAt.UTC().Format(time.RFC3339)
			dto.ExpiresAt = &s
		}
		brandIDs, err := h.getRestaurantBrandIDs(ctx, dto.ID)
		if err != nil {
			return nil, err
		}
		dto.RestaurantBrandIDs = brandIDs
		promos = append(promos, dto)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if promos == nil {
		promos = []PromoCodeDTO{}
	}
	return promos, nil
}

func (h *PromoHandler) getUserPromos(ctx context.Context, userID int64) ([]PromoCodeDTO, error) {
	// Промокоды, которые пользователь уже израсходовал (число использований
	// достигло max_uses), в список не попадают — повторно применить их нельзя.
	rows, err := h.pool.Query(ctx, `
		SELECT p.id, p.code, p.title, p.discount_percent, p.discount_amount,
		       p.min_order_amount, p.expires_at
		FROM user_promocode up
		JOIN promocode p ON p.id = up.promocode_id
		WHERE up.user_id = $1
		  AND (p.expires_at IS NULL OR p.expires_at > NOW())
		  AND (
		      SELECT COUNT(*) FROM promocode_usage pu
		      WHERE pu.promocode_id = p.id AND pu.client_account_id = $1
		  ) < COALESCE(p.max_uses, 1)
		ORDER BY up.added_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	return h.scanPromoList(ctx, rows)
}

func (h *PromoHandler) getRestaurantPromos(ctx context.Context, brandID int64) ([]PromoCodeDTO, error) {
	rows, err := h.pool.Query(ctx, `
		SELECT p.id, p.code, p.title, p.discount_percent, p.discount_amount,
		       p.min_order_amount, p.expires_at
		FROM promocode p
		JOIN promocode_restaurant_brand prb ON prb.promocode_id = p.id
		WHERE prb.restaurant_brand_id = $1
		  AND (p.expires_at IS NULL OR p.expires_at > NOW())
		ORDER BY p.created_at DESC
	`, brandID)
	if err != nil {
		return nil, err
	}
	return h.scanPromoList(ctx, rows)
}

func (h *PromoHandler) getRestaurantBrandIDs(ctx context.Context, promoID int64) ([]int64, error) {
	rows, err := h.pool.Query(ctx, `
		SELECT restaurant_brand_id FROM promocode_restaurant_brand
		WHERE promocode_id = $1
	`, promoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if ids == nil {
		ids = []int64{}
	}
	return ids, rows.Err()
}

func (h *PromoHandler) getByCode(ctx context.Context, code string) (*PromoCodeDTO, error) {
	var dto PromoCodeDTO
	var expiresAt *time.Time

	err := h.pool.QueryRow(ctx, `
		SELECT id, code, title, discount_percent, discount_amount,
		       min_order_amount, expires_at
		FROM promocode
		WHERE code = $1
	`, code).Scan(&dto.ID, &dto.Code, &dto.Title, &dto.DiscountPercent,
		&dto.DiscountAmount, &dto.MinOrderAmount, &expiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPromoNotFound
		}
		return nil, err
	}

	if expiresAt != nil {
		s := expiresAt.UTC().Format(time.RFC3339)
		dto.ExpiresAt = &s
	}

	brandIDs, err := h.getRestaurantBrandIDs(ctx, dto.ID)
	if err != nil {
		return nil, err
	}
	dto.RestaurantBrandIDs = brandIDs

	return &dto, nil
}

func (h *PromoHandler) bindPromo(ctx context.Context, userID int64, code string) (*PromoCodeDTO, error) {
	promo, err := h.getByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	if promo.ExpiresAt != nil {
		t, _ := time.Parse(time.RFC3339, *promo.ExpiresAt)
		if time.Now().UTC().After(t) {
			return nil, ErrPromoExpired
		}
	}

	tag, err := h.pool.Exec(ctx, `
		INSERT INTO user_promocode (user_id, promocode_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, userID, promo.ID)
	if err != nil {
		return nil, err
	}

	if tag.RowsAffected() == 0 {
		return nil, ErrAlreadyBound
	}

	return promo, nil
}

func (h *PromoHandler) validatePromo(ctx context.Context, userID int64, code string, brandID int64, orderAmount int64, deliveryCost int64, serviceFee int64) (*ValidatePromoResponse, error) {
	promo, err := h.getByCode(ctx, code)
	if err != nil {
		if errors.Is(err, ErrPromoNotFound) {
			return &ValidatePromoResponse{Valid: false, Reason: "promo not found"}, nil
		}
		return nil, err
	}

	if promo.ExpiresAt != nil {
		t, _ := time.Parse(time.RFC3339, *promo.ExpiresAt)
		if time.Now().UTC().After(t) {
			return &ValidatePromoResponse{Valid: false, Reason: "promo has expired"}, nil
		}
	}

	// Минимальная сумма заказа сверяется с учётом доставки и сервисного сбора.
	totalAmount := orderAmount + deliveryCost + serviceFee
	if totalAmount < promo.MinOrderAmount {
		return &ValidatePromoResponse{Valid: false, Reason: "order amount is below minimum"}, nil
	}

	if len(promo.RestaurantBrandIDs) > 0 && brandID > 0 {
		found := false
		for _, id := range promo.RestaurantBrandIDs {
			if id == brandID {
				found = true
				break
			}
		}
		if !found {
			return &ValidatePromoResponse{Valid: false, Reason: "promo is not valid for this restaurant"}, nil
		}
	}

	var usageCount int
	err = h.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM promocode_usage
		WHERE promocode_id = $1 AND client_account_id = $2
	`, promo.ID, userID).Scan(&usageCount)
	if err != nil {
		return nil, err
	}

	var maxUses int32
	err = h.pool.QueryRow(ctx, `
		SELECT COALESCE(max_uses, 1) FROM promocode WHERE id = $1
	`, promo.ID).Scan(&maxUses)
	if err != nil {
		return nil, err
	}

	if int32(usageCount) >= maxUses {
		return &ValidatePromoResponse{Valid: false, Reason: "max uses reached"}, nil
	}

	var discount int64
	if promo.DiscountAmount != nil {
		discount = *promo.DiscountAmount
	} else if promo.DiscountPercent != nil {
		discount = orderAmount * int64(*promo.DiscountPercent) / 100
	}

	// Скидка не может быть больше стоимости самих блюд.
	if discount > orderAmount {
		discount = orderAmount
	}

	return &ValidatePromoResponse{
		Valid:    true,
		Discount: discount,
	}, nil
}

func (h *PromoHandler) recordPromoUsage(ctx context.Context, userID int64, code string, orderPublicID string) error {
	promo, err := h.getByCode(ctx, code)
	if err != nil {
		return err
	}

	var usageCount int
	err = h.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM promocode_usage
		WHERE promocode_id = $1 AND client_account_id = $2
	`, promo.ID, userID).Scan(&usageCount)
	if err != nil {
		return err
	}

	var maxUses int32
	err = h.pool.QueryRow(ctx, `
		SELECT COALESCE(max_uses, 1) FROM promocode WHERE id = $1
	`, promo.ID).Scan(&maxUses)
	if err != nil {
		return err
	}

	if int32(usageCount) >= maxUses {
		return ErrMaxUsesReached
	}

	// Резолвим внутренний order.id по публичному id. Если заказ не найден,
	// пишем использование без привязки к заказу (order_id = NULL) — это
	// допустимо схемой и не роняет вставку из-за FK.
	var orderID *int64
	if orderPublicID != "" {
		var id int64
		if qerr := h.pool.QueryRow(ctx,
			`SELECT id FROM "order" WHERE public_id = $1`, orderPublicID).Scan(&id); qerr == nil {
			orderID = &id
		}
	}

	// ON CONFLICT по (promocode_id, order_id) защищает от повторной записи
	// одного и того же заказа; при этом разные заказы расходуют max_uses.
	_, err = h.pool.Exec(ctx, `
		INSERT INTO promocode_usage (promocode_id, client_account_id, order_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (promocode_id, order_id) DO NOTHING
	`, promo.ID, userID, orderID)
	return err
}
