package address

//go:generate easyjson $GOFILE

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase"
	address "github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/address"
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

func mapAddressToResponse(a domain.Address) AddressResponse {
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

type addressHandler struct {
	usecase address.AddressUseCase
	logger  domain.Logger
}

func NewAddressHandler(u usecase.AddressUseCase, l domain.Logger) *addressHandler {
	return &addressHandler{usecase: u, logger: l}
}

// AddAddress godoc
// @Summary 		Добавление нового адреса
// @Description		Создает новую запись адреса для текущего пользователя
// @Tags			profile
// @Accept			json
// @Produce			json
// @Param			input	body	  AddressRequest	true	"Данные адреса"
// @Success			201		{object}  map[string]string			"Успешное создание (возвращает public_id)"
// @Failure			400		{object}  response.ErrorResponse	"Ошибка в формате запроса"
// @Failure			401		{object}  response.ErrorResponse	"Неавторизован"
// @Failure			500		{object}  response.ErrorResponse	"Внутренняя ошибка сервера"
// @Router			/profile/addresses [post]
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

// GetAddresses godoc
// @Summary 		Получение списка адресов
// @Description		Возвращает все сохраненные адреса текущего пользователя
// @Tags			profile
// @Produce			json
// @Success			200		{object}  map[string][]AddressResponse	"Список адресов пользователя"
// @Failure			401		{object}  response.ErrorResponse		"Неавторизован"
// @Failure			500		{object}  response.ErrorResponse		"Внутренняя ошибка сервера"
// @Router			/profile/addresses [get]
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

	resp := make([]AddressResponse, 0, len(addresses))
	for _, addr := range addresses {
		resp = append(resp, mapAddressToResponse(addr))
	}

	response.JSON(w, http.StatusOK, map[string]any{"addresses": resp})
}

// DeleteAddress godoc
// @Summary 		Удаление адреса
// @Description		Удаляет адрес пользователя по его ID
// @Tags			profile
// @Param			id		path	  string	true	"Public ID адреса"
// @Success			200		{object}  map[string]string			"Адрес успешно удален"
// @Failure			401		{object}  response.ErrorResponse	"Неавторизован"
// @Failure			500		{object}  response.ErrorResponse	"Внутренняя ошибка сервера"
// @Router			/profile/addresses/{id} [delete]
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

// UpdateAddress godoc
// @Summary 		Обновление адреса
// @Description		Изменяет данные существующего адреса пользователя
// @Tags			profile
// @Accept			json
// @Produce			json
// @Param			id		path	  string			true	"Public ID адреса"
// @Param			input	body	  AddressRequest	true	"Новые данные адреса"
// @Success			200		{object}  map[string]string			"Адрес успешно обновлен"
// @Failure			400		{object}  response.ErrorResponse	"Ошибка в формате запроса"
// @Failure			401		{object}  response.ErrorResponse	"Неавторизован"
// @Failure			500		{object}  response.ErrorResponse	"Внутренняя ошибка сервера"
// @Router			/profile/addresses/{id} [patch]
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
