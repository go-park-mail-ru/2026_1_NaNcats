package handler

//go:generate easyjson $GOFILE

import (
	"net/http"
	"strconv"

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
	userID, _ := middleware.GetUserID(ctx)

	var req AddressRequest
	if err := request.JSON(r, &req); err != nil {
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
		h.logger.Error("failed to add address", err, nil)
		response.Error(w, http.StatusInternalServerError, "failed to save address")
		return
	}

	response.JSON(w, http.StatusCreated, map[string]int{"id": id})
}

func (h *addressHandler) GetAddresses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := middleware.GetUserID(ctx)

	addresses, err := h.usecase.GetMyAddresses(ctx, userID)
	if err != nil {
		h.logger.Error("failed to get addresses", err, nil)
		response.Error(w, http.StatusInternalServerError, "failed to fetch addresses")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{"addresses": addresses})
}

func (h *addressHandler) DeleteAddress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := middleware.GetUserID(ctx)

	idStr := r.PathValue("id")
	addrID, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid address id")
		return
	}

	if err := h.usecase.DeleteAddress(ctx, userID, addrID); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to delete address")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}
