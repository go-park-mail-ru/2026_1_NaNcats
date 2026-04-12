package handler

//go:generate easyjson $GOFILE

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/request"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/response"
)

//easyjson:json
type AddressRequest struct {
	AddressText    string  `json:"address_text"`
	Lat            float64 `json:"lat"`
	Lon            float64 `json:"lon"`
	Apartment      string  `json:"apartment"`
	Entrance       string  `json:"entrance"`
	Floor          string  `json:"floor"`
	DoorCode       string  `json:"door_code"`
	CourierComment string  `json:"courier_comment"`
	Label          string  `json:"label"`
}

type addressHandler struct {
	usecase usecase.AddressUseCase
	logger  domain.Logger
}

func NewAddressHandler(u usecase.AddressUseCase, l domain.Logger) *addressHandler {
	return &addressHandler{usecase: u, logger: l}
}

func (h *addressHandler) AddAddress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		l.Error("failed to get user_id from context", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	var req AddressRequest
	if err := request.JSON(r, &req); err != nil {
		l.Warn("failed to decode add address request", domain.Int("user_id", userID), domain.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	addr := domain.Address{
		Location: domain.Location{
			AddressText: req.AddressText,
			Latitude:    req.Lat,
			Longitude:   req.Lon,
		},
		Apartment:      req.Apartment,
		Entrance:       req.Entrance,
		Floor:          req.Floor,
		DoorCode:       req.DoorCode,
		CourierComment: req.CourierComment,
		Label:          req.Label,
	}

	id, err := h.usecase.AddAddress(ctx, userID, addr)
	if err != nil {
		l.Error("failed to add address to database", err, domain.Int("user_id", userID))
		response.Error(w, http.StatusInternalServerError, "failed to save address")
		return
	}

	l.Info("address added successfully", domain.Int("user_id", userID), domain.String("address_id", id))

	response.JSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *addressHandler) GetAddresses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		l.Error("failed to get user_id from context", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	addresses, err := h.usecase.GetMyAddresses(ctx, userID)
	if err != nil {
		l.Error("failed to fetch addresses", err, domain.Int("user_id", userID))
		response.Error(w, http.StatusInternalServerError, "failed to fetch addresses")
		return
	}

	l.Debug("fetched user addresses", domain.Int("user_id", userID), domain.Int("count", len(addresses)))

	response.JSON(w, http.StatusOK, map[string]any{"addresses": addresses})
}

func (h *addressHandler) DeleteAddress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		l.Error("failed to get user_id from context", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	addressPublicID := r.PathValue("id")
	if err := h.usecase.DeleteAddress(ctx, userID, addressPublicID); err != nil {
		l.Error("failed to delete address", err, domain.Int("user_id", userID), domain.String("address_id", addressPublicID))
		response.Error(w, http.StatusInternalServerError, "failed to delete address")
		return
	}

	l.Info("address deleted successfully", domain.Int("user_id", userID), domain.String("address_id", addressPublicID))

	response.JSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

func (h *addressHandler) UpdateAddress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		l.Error("failed to get user_id from context", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	idStr := r.PathValue("id")

	var req AddressRequest
	if err := request.JSON(r, &req); err != nil {
		l.Warn("failed to decode update address request", domain.Int("user_id", userID), domain.String("address_id", idStr), domain.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	addr := domain.Address{
		PublicID: idStr,
		Location: domain.Location{
			AddressText: req.AddressText,
			Latitude:    req.Lat,
			Longitude:   req.Lon,
		},
		Apartment:      req.Apartment,
		Entrance:       req.Entrance,
		Floor:          req.Floor,
		DoorCode:       req.DoorCode,
		CourierComment: req.CourierComment,
		Label:          req.Label,
	}

	err = h.usecase.UpdateAddress(ctx, userID, addr)
	if err != nil {
		l.Error("failed to update address in database", err, domain.Int("user_id", userID), domain.String("address_id", idStr))
		response.Error(w, http.StatusInternalServerError, "failed to update address")
		return
	}

	l.Info("address updated successfully", domain.Int("user_id", userID), domain.String("address_id", idStr))

	response.JSON(w, http.StatusOK, map[string]string{"message": "address updated successfully"})
}
