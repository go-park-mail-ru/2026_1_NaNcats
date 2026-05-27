package promo

//go:generate easyjson $GOFILE

import (
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/request"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/response"
	pbOrder "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/order"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	Code          string `json:"code"`
	OrderPublicID string `json:"order_public_id"`
}

type PromoHandler struct {
	client pbOrder.PromoServiceClient
	logger logger.Logger
}

func NewPromoHandler(client pbOrder.PromoServiceClient, l logger.Logger) *PromoHandler {
	return &PromoHandler{client: client, logger: l}
}

func toDTO(p *pbOrder.Promocode) PromoCodeDTO {
	ids := p.GetRestaurantBrandIds()
	if ids == nil {
		ids = []int64{}
	}
	dto := PromoCodeDTO{
		ID:                 p.GetId(),
		Code:               p.GetCode(),
		Title:              p.GetTitle(),
		MinOrderAmount:     p.GetMinOrderAmount(),
		RestaurantBrandIDs: ids,
	}
	if p.DiscountPercent != nil {
		dto.DiscountPercent = p.DiscountPercent
	}
	if p.DiscountAmount != nil {
		dto.DiscountAmount = p.DiscountAmount
	}
	if p.ExpiresAt != nil {
		dto.ExpiresAt = p.ExpiresAt
	}
	return dto
}

func toDTOList(list *pbOrder.PromocodeList) []PromoCodeDTO {
	out := make([]PromoCodeDTO, 0, len(list.GetPromocodes()))
	for _, p := range list.GetPromocodes() {
		out = append(out, toDTO(p))
	}
	return out
}

func (h *PromoHandler) GetUserPromos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	resp, err := h.client.GetUserPromos(ctx, &pbOrder.GetUserPromosRequest{UserId: userID})
	if err != nil {
		l.Error("failed to get user promos", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, toDTOList(resp))
}

func (h *PromoHandler) GetRestaurantPromos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	brandID, err := strconv.ParseInt(r.URL.Query().Get("brand_id"), 10, 64)
	if err != nil || brandID <= 0 {
		response.Error(w, http.StatusBadRequest, "brand_id is required")
		return
	}

	resp, err := h.client.GetRestaurantPromos(ctx, &pbOrder.GetRestaurantPromosRequest{RestaurantBrandId: brandID})
	if err != nil {
		l.Error("failed to get restaurant promos", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, toDTOList(resp))
}

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
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Code == "" {
		response.Error(w, http.StatusBadRequest, "code is required")
		return
	}

	promo, err := h.client.BindPromocode(ctx, &pbOrder.BindPromocodeRequest{
		UserId: userID,
		Code:   req.Code,
	})
	if err != nil {
		h.writePromoError(w, l, "failed to bind promo", err)
		return
	}

	response.JSON(w, http.StatusOK, toDTO(promo))
}

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
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Code == "" {
		response.Error(w, http.StatusBadRequest, "code is required")
		return
	}

	resp, err := h.client.ValidatePromocode(ctx, &pbOrder.ValidatePromocodeRequest{
		UserId:            userID,
		Code:              req.Code,
		RestaurantBrandId: req.RestaurantBrandID,
		OrderAmount:       req.OrderAmount,
		DeliveryCost:      req.DeliveryCost,
		ServiceFee:        req.ServiceFee,
	})
	if err != nil {
		l.Error("failed to validate promo", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, ValidatePromoResponse{
		Valid:    resp.GetValid(),
		Discount: resp.GetDiscount(),
		Reason:   resp.GetReason(),
	})
}

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
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Code == "" {
		response.Error(w, http.StatusBadRequest, "code is required")
		return
	}

	_, err := h.client.UsePromocode(ctx, &pbOrder.UsePromocodeRequest{
		UserId:        userID,
		Code:          req.Code,
		OrderPublicId: req.OrderPublicID,
	})
	if err != nil {
		h.writePromoError(w, l, "failed to record promo usage", err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *PromoHandler) writePromoError(w http.ResponseWriter, l logger.Logger, logMsg string, err error) {
	st, ok := status.FromError(err)
	if !ok {
		l.Error(logMsg, err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	switch st.Code() {
	case codes.NotFound:
		response.Error(w, http.StatusBadRequest, "promo not found")
	case codes.AlreadyExists:
		response.Error(w, http.StatusConflict, "already bound")
	case codes.FailedPrecondition:
		response.Error(w, http.StatusBadRequest, st.Message())
	default:
		l.Error(logMsg, err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
	}
}
