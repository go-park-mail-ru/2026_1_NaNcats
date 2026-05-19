package cart

//go:generate easyjson $GOFILE

import (
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/delivery/websocket"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/cartclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/userclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/request"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/response"
	pbUser "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/user"
	gorillaWs "github.com/gorilla/websocket"
)

var upgrader = gorillaWs.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

//easyjson:json
type CartItemDTO struct {
	DishID        int64   `json:"dish_id"`
	Name          string  `json:"name,omitempty"`
	Price         int64   `json:"price,omitempty"`
	Quantity      int32   `json:"quantity"`
	ImageURL      string  `json:"image_url,omitempty"`
	OwnerPublicID *string `json:"owner_public_id,omitempty"`
	OwnerName     *string `json:"owner_name,omitempty"`
	OwnerAvatar   *string `json:"owner_avatar,omitempty"`
}

//easyjson:json
type CartMemberDTO struct {
	PublicID  string `json:"public_id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	JoinedAt  string `json:"joined_at"`
}

//easyjson:json
type CartResponse struct {
	CartID            string          `json:"cart_id"`
	AdminID           string          `json:"admin_id"`
	RestaurantBrandID int64           `json:"restaurant_id"`
	Mode              string          `json:"mode"`
	Status            string          `json:"status"`
	TotalCost         int64           `json:"total_cost"`
	Items             []CartItemDTO   `json:"items"`
	Members           []CartMemberDTO `json:"members,omitempty"` // Только для shared
}

//easyjson:json
type AddItemRequest struct {
	CartID   string `json:"cart_id"`
	DishID   int64  `json:"dish_id"`
	Quantity int32  `json:"quantity"`
}

//easyjson:json
type InviteResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

//easyjson:json
type RemoveItemRequest struct {
	CartID string `json:"cart_id"`
	DishID int64  `json:"dish_id"`
}

//easyjson:json
type UpdateQuantityRequest struct {
	CartID   string `json:"cart_id"`
	DishID   int64  `json:"dish_id"`
	Quantity int32  `json:"quantity"`
}

//easyjson:json
type ReassignOwnerRequest struct {
	CartID           string  `json:"cart_id"`
	DishID           int64   `json:"dish_id"`
	NewOwnerPublicID *string `json:"new_owner_public_id"`
}

//easyjson:json
type JoinCartRequest struct {
	Token string `json:"token"`
}

//easyjson:json
type KickMemberRequest struct {
	CartID         string `json:"cart_id"`
	TargetPublicID string `json:"target_public_id"`
}

//easyjson:json
type BasicCartOperationRequest struct {
	CartID string `json:"cart_id"`
}

type CartHandler struct {
	cartClient cartclient.CartClient
	userClient userclient.UserClient
	wsManager  *websocket.WsManager
	logger     logger.Logger
}

func NewCartHandler(cc cartclient.CartClient, uc userclient.UserClient, wm *websocket.WsManager, l logger.Logger) *CartHandler {
	return &CartHandler{
		cartClient: cc,
		userClient: uc,
		wsManager:  wm,
		logger:     l,
	}
}

// GetCart godoc
// @Summary      Получить корзину
// @Description  Возвращает текущую корзину авторизованного пользователя (включая статус комнаты и участников)
// @Tags         cart
// @Produce      json
// @Success      200  {object}  CartResponse
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Security     ApiKeyAuth
// @Router       /api/cart [get]
func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	cart, err := h.cartClient.GetCart(ctx, userID)
	if err != nil {
		h.logger.Error("failed to fetch cart via grpc", err)
		response.Error(w, http.StatusInternalServerError, "failed to get cart")
		return
	}

	userIDsSet := make(map[int64]struct{})
	userIDsSet[cart.AdminID] = struct{}{}
	for _, item := range cart.Items {
		if item.OwnerUserID != nil {
			userIDsSet[*item.OwnerUserID] = struct{}{}
		}
	}
	for _, m := range cart.Members {
		userIDsSet[m.UserID] = struct{}{}
	}

	userIDs := make([]int64, 0, len(userIDsSet))
	for id := range userIDsSet {
		userIDs = append(userIDs, id)
	}

	usersInfo, err := h.userClient.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		h.logger.Warn("failed to fetch users info for enrichment, returning partial data", logger.Err(err))
		usersInfo = make(map[int64]*pbUser.User)
	}

	adminPublicID := "unknown"
	if u, ok := usersInfo[cart.AdminID]; ok {
		adminPublicID = u.PublicId
	}

	cartResponse := CartResponse{
		CartID:            cart.ID,
		AdminID:           adminPublicID,
		RestaurantBrandID: cart.RestaurantBrandID,
		Mode:              cart.Mode,
		Status:            cart.Status,
		TotalCost:         cart.TotalCost,
		Items:             make([]CartItemDTO, 0, len(cart.Items)),
		Members:           make([]CartMemberDTO, 0, len(cart.Members)),
	}

	for _, item := range cart.Items {
		var pubID, name, avatar *string
		if item.OwnerUserID != nil {
			if u, exists := usersInfo[*item.OwnerUserID]; exists {
				pubID = &u.PublicId
				name = &u.Name
				avatar = &u.AvatarUrl
			}
		}

		cartResponse.Items = append(cartResponse.Items, CartItemDTO{
			DishID:        item.DishID,
			Name:          item.Name,
			Price:         item.Price,
			Quantity:      item.Quantity,
			ImageURL:      item.ImageURL,
			OwnerPublicID: pubID,
			OwnerName:     name,
			OwnerAvatar:   avatar,
		})
	}

	for _, m := range cart.Members {
		var pubID, name, avatar string
		if u, exists := usersInfo[m.UserID]; exists {
			pubID = u.PublicId
			name = u.Name
			avatar = u.AvatarUrl
		} else {
			pubID = "unknown"
			name = "Unknown User"
		}

		cartResponse.Members = append(cartResponse.Members, CartMemberDTO{
			PublicID:  pubID,
			Name:      name,
			AvatarURL: avatar,
			JoinedAt:  m.JoinedAt,
		})
	}

	response.JSON(w, http.StatusOK, cartResponse)
}

// ConnectCartWS godoc
// @Summary      WebSocket соединение корзины
// @Description  Устанавливает WS соединение для получения live-обновлений совместной корзины
// @Tags         cart
// @Param        cart_id  query  string  true  "ID корзины"
// @Success      101  "Switching Protocols"
// @Failure      400  {object}  map[string]string "Bad Request (отсутствует cart_id)"
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Security     ApiKeyAuth
// @Router       /api/ws/cart [get]
func (h *CartHandler) ConnectCartWS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	cartID := r.URL.Query().Get("cart_id")
	if cartID == "" {
		response.Error(w, http.StatusBadRequest, "cart_id is required")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("failed to upgrade connection to websocket", err, logger.Int("user_id", int(userID)))
		return
	}

	h.wsManager.AddCartConnection(cartID, userID, conn)
}

// AddItem godoc
// @Summary      Добавить товар
// @Description  Добавляет товар в корзину. Вызывающий пользователь становится владельцем (owner) товара.
// @Tags         cart
// @Accept       json
// @Produce      json
// @Param        Idempotency-Key  header    string          true  "Ключ идемпотентности"
// @Param        input            body      AddItemRequest  true  "Данные товара"
// @Success      200  {object}  map[string]string "Товар добавлен"
// @Failure      400  {object}  map[string]string "Неверное количество или блюдо из другого ресторана"
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Failure      403  {object}  map[string]string "Пользователь не состоит в этой корзине"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Security     ApiKeyAuth
// @Router       /api/cart/items [post]
func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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

	var req AddItemRequest
	if err := request.JSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.cartClient.AddItem(ctx, req.CartID, userID, req.DishID, req.Quantity, idemKey)
	if err != nil {
		if errors.Is(err, cartclient.ErrForbidden) {
			response.Error(w, http.StatusForbidden, "you have no rights to add items to this cart")
			return
		}
		if errors.Is(err, cartclient.ErrMultipleRestaurants) {
			response.Error(w, http.StatusConflict, "MULTIPLE_RESTAURANTS")
			return
		}
		if errors.Is(err, cartclient.ErrCartLocked) {
			response.Error(w, http.StatusConflict, "CART_LOCKED")
			return
		}
		if errors.Is(err, cartclient.ErrInvalidQuantity) {
			response.Error(w, http.StatusBadRequest, "invalid quantity")
			return
		}
		if errors.Is(err, cartclient.ErrInvalidCart) {
			response.Error(w, http.StatusBadRequest, "invalid cart data")
			return
		}

		h.logger.Error("failed to add item via grpc", err)
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "item added"})
}

// GenerateInvite godoc
// @Summary      Сгенерировать инвайт-ссылку
// @Description  Создает токен-приглашение в корзину. Переводит корзину в режим 'shared'. Доступно только админу.
// @Tags         cart
// @Produce      json
// @Param        cart_id          query     string  true  "ID корзины"
// @Success      200  {object}  InviteResponse
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Failure      403  {object}  map[string]string "Только админ может генерировать ссылки"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Security     ApiKeyAuth
// @Router       /api/cart/invite [post]
func (h *CartHandler) GenerateInvite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	cartID := r.URL.Query().Get("cart_id")

	invite, err := h.cartClient.GenerateInvite(ctx, cartID, userID)
	if err != nil {
		if errors.Is(err, cartclient.ErrForbidden) {
			response.Error(w, http.StatusForbidden, "only admin can generate invites")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to generate invite")
		return
	}

	response.JSON(w, http.StatusOK, InviteResponse{
		Token:     invite.Token,
		ExpiresAt: invite.ExpiresAt,
	})
}

// RemoveItem godoc
// @Summary      Удалить товар
// @Description  Удаляет позицию из корзины. Гость может удалять только свое, админ - любое.
// @Tags         cart
// @Accept       json
// @Produce      json
// @Param        Idempotency-Key  header    string             true  "Ключ идемпотентности"
// @Param        input            body      RemoveItemRequest  true  "Данные для удаления"
// @Success      200  {object}  map[string]string "Товар удален"
// @Failure      400  {object}  map[string]string "Неверный формат запроса"
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Failure      403  {object}  map[string]string "Нет прав на удаление этой позиции"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Security     ApiKeyAuth
// @Router       /api/cart/items [delete]
func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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

	var req RemoveItemRequest
	if err := request.JSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.cartClient.RemoveItem(ctx, req.CartID, userID, req.DishID, idemKey)
	if err != nil {
		if errors.Is(err, cartclient.ErrForbidden) {
			response.Error(w, http.StatusForbidden, "you can only remove your own items")
			return
		}
		h.logger.Error("failed to remove item", err)
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "item removed"})
}

// UpdateQuantity godoc
// @Summary      Изменить количество товара
// @Description  Изменяет количество существующего товара. Гость может менять только свои позиции, админ - любые.
// @Tags         cart
// @Accept       json
// @Produce      json
// @Param        Idempotency-Key  header    string                 true  "Ключ идемпотентности"
// @Param        input            body      UpdateQuantityRequest  true  "Данные обновления"
// @Success      200  {object}  map[string]string "Количество обновлено"
// @Failure      400  {object}  map[string]string "Неверный запрос или количество"
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Failure      403  {object}  map[string]string "Нет прав на изменение этой позиции"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Security     ApiKeyAuth
// @Router       /api/cart/items [put]
func (h *CartHandler) UpdateQuantity(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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

	var req UpdateQuantityRequest
	if err := request.JSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.cartClient.UpdateItemQuantity(ctx, req.CartID, userID, req.DishID, req.Quantity, idemKey)
	if err != nil {
		if errors.Is(err, cartclient.ErrInvalidCart) {
			response.Error(w, http.StatusBadRequest, "invalid quantity")
			return
		}
		if errors.Is(err, cartclient.ErrForbidden) {
			response.Error(w, http.StatusForbidden, "you can only update your own items")
			return
		}
		h.logger.Error("failed to update quantity", err)
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "quantity updated"})
}

// ReassignOwner godoc
// @Summary      Изменить плательщика товара
// @Description  Переназначает блюдо на другого участника комнаты (или делает позицию ничейной, передав null). Только для админа.
// @Tags         cart
// @Accept       json
// @Produce      json
// @Param        Idempotency-Key  header    string                true  "Ключ идемпотентности"
// @Param        input            body      ReassignOwnerRequest  true  "Новый владелец"
// @Success      200  {object}  map[string]string "Владелец изменен"
// @Failure      400  {object}  map[string]string "Неверный формат запроса"
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Failure      403  {object}  map[string]string "Только админ может переназначать"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Security     ApiKeyAuth
// @Router       /api/cart/items/owner [patch]
func (h *CartHandler) ReassignOwner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		response.Error(w, http.StatusBadRequest, "Idempotency-Key is required")
		return
	}

	var req ReassignOwnerRequest
	if err := request.JSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var newOwnerID *int64
	if req.NewOwnerPublicID != nil {
		id, err := h.userClient.ResolvePublicID(ctx, *req.NewOwnerPublicID)
		if err != nil {
			if errors.Is(err, userclient.ErrUserNotFound) {
				response.Error(w, http.StatusBadRequest, "target user not found")
				return
			}
			h.logger.Error("failed to resolve public ID", err)
			response.Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
		newOwnerID = &id
	}

	err := h.cartClient.ReassignItemOwner(ctx, req.CartID, userID, req.DishID, newOwnerID, idemKey)
	if err != nil {
		if errors.Is(err, cartclient.ErrForbidden) {
			response.Error(w, http.StatusForbidden, "only admin can reassign items")
			return
		}
		h.logger.Error("failed to reassign owner", err)
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "owner reassigned"})
}

// JoinCart godoc
// @Summary      Присоединиться к корзине
// @Description  Добавляет пользователя в совместную корзину по токену инвайта.
// @Tags         cart
// @Accept       json
// @Produce      json
// @Param        input  body      JoinCartRequest  true  "Токен приглашения"
// @Success      200    {object}  map[string]string "Успешный вход (возвращает cart_id)"
// @Failure      400    {object}  map[string]string "Неверный формат запроса"
// @Failure      401    {object}  map[string]string "Unauthorized"
// @Failure      404    {object}  map[string]string "Инвайт недействителен или протух"
// @Failure      500    {object}  map[string]string "Internal server error"
// @Security     ApiKeyAuth
// @Router       /api/cart/join [post]
func (h *CartHandler) JoinCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req JoinCartRequest
	if err := request.JSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cartID, err := h.cartClient.JoinCart(ctx, req.Token, userID)
	if err != nil {
		if errors.Is(err, cartclient.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "invite link is invalid or expired")
			return
		}
		h.logger.Error("failed to join cart", err)
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"cart_id": cartID})
}

// KickMember godoc
// @Summary      Удалить участника
// @Description  Кикает гостя из комнаты. Блюда гостя становятся 'ничейными' (owner_user_id = null).
// @Tags         cart
// @Accept       json
// @Produce      json
// @Param        Idempotency-Key  header    string             true  "Ключ идемпотентности"
// @Param        input            body      KickMemberRequest  true  "Кого удаляем"
// @Success      200  {object}  map[string]string "Участник удален"
// @Failure      400  {object}  map[string]string "Неверный формат запроса"
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Failure      403  {object}  map[string]string "Только админ может удалять участников"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Security     ApiKeyAuth
// @Router       /api/cart/members [delete]
func (h *CartHandler) KickMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		response.Error(w, http.StatusBadRequest, "Idempotency-Key is required")
		return
	}

	var req KickMemberRequest
	if err := request.JSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	targetInternalID, err := h.userClient.ResolvePublicID(ctx, req.TargetPublicID)
	if err != nil {
		if errors.Is(err, userclient.ErrUserNotFound) {
			response.Error(w, http.StatusBadRequest, "target user not found")
			return
		}
		h.logger.Error("failed to resolve public ID", err)
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	err = h.cartClient.KickMember(ctx, req.CartID, userID, targetInternalID, idemKey)
	if err != nil {
		if errors.Is(err, cartclient.ErrForbidden) {
			response.Error(w, http.StatusForbidden, "only admin can kick members")
			return
		}
		h.logger.Error("failed to kick member", err)
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "member kicked"})
}

// CloseSharedCart godoc
// @Summary      Закрыть совместную корзину
// @Description  Переводит корзину в Solo-режим. Удаляет всех участников кроме админа и удаляет их блюда.
// @Tags         cart
// @Accept       json
// @Produce      json
// @Param        Idempotency-Key  header    string                     true  "Ключ идемпотентности"
// @Param        input            body      BasicCartOperationRequest  true  "ID корзины"
// @Success      200  {object}  map[string]string "Корзина переведена в соло-режим"
// @Failure      400  {object}  map[string]string "Неверный формат запроса"
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Failure      403  {object}  map[string]string "Только админ может закрыть корзину"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Security     ApiKeyAuth
// @Router       /api/cart/close [post]
func (h *CartHandler) CloseSharedCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		response.Error(w, http.StatusBadRequest, "Idempotency-Key is required")
		return
	}

	var req BasicCartOperationRequest
	if err := request.JSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.cartClient.CloseSharedCart(ctx, req.CartID, userID, idemKey)
	if err != nil {
		if errors.Is(err, cartclient.ErrForbidden) {
			response.Error(w, http.StatusForbidden, "only admin can close shared cart")
			return
		}
		h.logger.Error("failed to close shared cart", err)
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "cart is now solo"})
}

// ClearCart godoc
// @Summary      Очистить корзину
// @Description  Удаляет все товары из указанной корзины. В Shared-режиме доступно только админу.
// @Tags         cart
// @Accept       json
// @Produce      json
// @Param        Idempotency-Key  header    string                     true  "Ключ идемпотентности"
// @Param        input            body      BasicCartOperationRequest  true  "Данные"
// @Success      200  {object}  map[string]string "Корзина очищена"
// @Failure      400  {object}  map[string]string "Неверный формат запроса"
// @Failure      401  {object}  map[string]string "Unauthorized"
// @Failure      403  {object}  map[string]string "Нет прав на выполнение операции"
// @Failure      500  {object}  map[string]string "Internal server error"
// @Security     ApiKeyAuth
// @Router       /api/cart [delete]
func (h *CartHandler) ClearCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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

	var req BasicCartOperationRequest
	if err := request.JSON(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.cartClient.ClearCart(ctx, req.CartID, userID, idemKey)
	if err != nil {
		if errors.Is(err, cartclient.ErrForbidden) {
			response.Error(w, http.StatusForbidden, "only admin can clear cart")
			return
		}
		h.logger.Error("failed to clear cart", err)
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "cart cleared"})
}
