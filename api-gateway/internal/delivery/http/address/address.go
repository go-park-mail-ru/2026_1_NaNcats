package address

//go:generate easyjson $GOFILE

import (
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/addressclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/request"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/response"
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

//easyjson:json
type LocationResponse struct {
	AddressText string  `json:"address_text"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}

//easyjson:json
type AddressResponse struct {
	ID             string           `json:"id"`
	Location       LocationResponse `json:"location"`
	Apartment      string           `json:"apartment"`
	Entrance       string           `json:"entrance"`
	Floor          string           `json:"floor"`
	DoorCode       string           `json:"door_code"`
	CourierComment string           `json:"courier_comment"`
	Label          string           `json:"label"`
}

//easyjson:json
type CreateAddressResponse struct {
	ID string `json:"id"`
}

//easyjson:json
type AddressListResponse struct {
	Addresses []AddressResponse `json:"addresses"`
}

//easyjson:json
type MessageResponse struct {
	Message string `json:"message"`
}

func mapAddressToResponse(a addressclient.Address) AddressResponse {
	return AddressResponse{
		ID: a.PublicID,
		Location: LocationResponse{
			AddressText: a.Location.AddressText,
			Latitude:    a.Location.Latitude,
			Longitude:   a.Location.Longitude,
		},
		Apartment:      a.Apartment,
		Entrance:       a.Entrance,
		Floor:          a.Floor,
		DoorCode:       a.DoorCode,
		CourierComment: a.CourierComment,
		Label:          a.Label,
	}
}

type AddressHandler struct {
	addressClient addressclient.AddressClient
	logger        logger.Logger
}

func NewAddressHandler(ac addressclient.AddressClient, l logger.Logger) *AddressHandler {
	return &AddressHandler{
		addressClient: ac,
		logger:        l,
	}
}

// AddAddress godoc
// @Summary 		Добавление нового адреса
// @Description		Создает новую запись адреса для текущего пользователя
// @Tags			profile
// @Accept			json
// @Produce			json
// @Param			input	body	  AddressRequest	true	"Данные адреса"
// @Success			201		{object}  CreateAddressResponse			"Успешное создание (возвращает public_id)"
// @Failure			400		{object}  response.ErrorResponse	"Ошибка в формате запроса"
// @Failure			401		{object}  response.ErrorResponse	"Неавторизован"
// @Failure			500		{object}  response.ErrorResponse	"Внутренняя ошибка сервера"
// @Router			/profile/addresses [post]
func (h *AddressHandler) AddAddress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		response.Error(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	var req AddressRequest
	if err := request.JSON(r, &req); err != nil {
		l.Warn("failed to decode add address request", logger.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	addr := addressclient.Address{
		Location: addressclient.Location{
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

	id, err := h.addressClient.AddAddress(ctx, userID, addr, idemKey)
	if err != nil {
		l.Error("failed to add address via grpc", err)
		response.Error(w, http.StatusInternalServerError, "failed to save address")
		return
	}

	response.JSON(w, http.StatusCreated, CreateAddressResponse{ID: id})
}

// GetAddresses godoc
// @Summary 		Получение списка адресов
// @Description		Возвращает все сохраненные адреса текущего пользователя
// @Tags			profile
// @Produce			json
// @Success			200		{object}  AddressListResponse	"Список адресов пользователя"
// @Failure			401		{object}  response.ErrorResponse		"Неавторизован"
// @Failure			500		{object}  response.ErrorResponse		"Внутренняя ошибка сервера"
// @Router			/profile/addresses [get]
func (h *AddressHandler) GetAddresses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	addresses, err := h.addressClient.GetMyAddresses(ctx, userID)
	if err != nil {
		l.Error("failed to fetch addresses via grpc", err)
		response.Error(w, http.StatusInternalServerError, "failed to fetch addresses")
		return
	}

	resp := make([]AddressResponse, 0, len(addresses))
	for _, addr := range addresses {
		resp = append(resp, mapAddressToResponse(addr))
	}

	response.JSON(w, http.StatusOK, AddressListResponse{Addresses: resp})
}

// DeleteAddress godoc
// @Summary 		Удаление адреса
// @Description		Удаляет адрес пользователя по его ID
// @Tags			profile
// @Param			id		path	  string	true	"Public ID адреса"
// @Success			200		{object}  MessageResponse			"Адрес успешно удален"
// @Failure			401		{object}  response.ErrorResponse	"Неавторизован"
// @Failure			500		{object}  response.ErrorResponse	"Внутренняя ошибка сервера"
// @Router			/profile/addresses/{id} [delete]
func (h *AddressHandler) DeleteAddress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		response.Error(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	addressPublicID := r.PathValue("id")
	if addressPublicID == "" {
		response.Error(w, http.StatusBadRequest, "Address ID is required")
		return
	}

	err := h.addressClient.DeleteAddress(ctx, userID, addressPublicID, idemKey)
	if err != nil {
		if errors.Is(err, addressclient.ErrAddressNotFound) {
			response.Error(w, http.StatusNotFound, "Address not found")
			return
		}
		l.Error("failed to delete address via grpc", err)
		response.Error(w, http.StatusInternalServerError, "failed to delete address")
		return
	}

	response.JSON(w, http.StatusOK, MessageResponse{Message: "deleted"})
}

// UpdateAddress godoc
// @Summary 		Обновление адреса
// @Description		Изменяет данные существующего адреса пользователя
// @Tags			profile
// @Accept			json
// @Produce			json
// @Param			id		path	  string			true	"Public ID адреса"
// @Param			input	body	  AddressRequest	true	"Новые данные адреса"
// @Success			200		{object}  MessageResponse			"Адрес успешно обновлен"
// @Failure			400		{object}  response.ErrorResponse	"Ошибка в формате запроса"
// @Failure			401		{object}  response.ErrorResponse	"Неавторизован"
// @Failure			500		{object}  response.ErrorResponse	"Внутренняя ошибка сервера"
// @Router			/profile/addresses/{id} [patch]
func (h *AddressHandler) UpdateAddress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		response.Error(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	addressPublicID := r.PathValue("id")
	if addressPublicID == "" {
		response.Error(w, http.StatusBadRequest, "Address ID is required")
		return
	}

	var req AddressRequest
	if err := request.JSON(r, &req); err != nil {
		l.Warn("failed to decode update address request", logger.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	addr := addressclient.Address{
		PublicID: addressPublicID,
		Location: addressclient.Location{
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

	err := h.addressClient.UpdateAddress(ctx, userID, addr, idemKey)
	if err != nil {
		if errors.Is(err, addressclient.ErrAddressNotFound) {
			response.Error(w, http.StatusNotFound, "Address not found")
			return
		}
		l.Error("failed to update address via grpc", err)
		response.Error(w, http.StatusInternalServerError, "failed to update address")
		return
	}

	response.JSON(w, http.StatusOK, MessageResponse{Message: "address updated successfully"})
}
