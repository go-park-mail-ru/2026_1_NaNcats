package wheel

//go:generate easyjson $GOFILE

import (
	"math/rand"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/orderclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/userclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/response"
	pbOrder "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/order"
)

// Описание секторов колеса, их эмодзи-иконок и весов
type Sector struct {
	ID     int
	Name   string
	Emoji  string
	Weight int // Общая сумма весов равна 1000 (100%)
}

//easyjson:json
type SectorResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Emoji string `json:"emoji"`
}

//easyjson:json
type WheelSectorsResponse struct {
	Sectors []SectorResponse `json:"sectors"`
}

var sectors = []Sector{
	{ID: 1, Name: "Попробуй в следующий раз", Emoji: "🍂", Weight: 300},   // 30% (ничего)
	{ID: 2, Name: "Персональная скидка", Emoji: "💸", Weight: 150},        // 15% (150 руб)
	{ID: 3, Name: "Скидка на популярный бренд", Emoji: "🔥", Weight: 150}, // 15% (15% скидка на бренд)
	{ID: 4, Name: "Заморозка стрика", Emoji: "❄️", Weight: 100},          // 10% (Streak Freeze)
	{ID: 5, Name: "Стрик-буст", Emoji: "🚀", Weight: 100},                 // 10% (+1 неделя)
	{ID: 6, Name: "Эксклюзивная ачивка", Emoji: "🎡", Weight: 20},         // 2%  (Редкая награда)
	{ID: 7, Name: "Супер-приз", Emoji: "🎁", Weight: 10},                  // 1%  (Крупный промокод)
	{ID: 8, Name: "Реролл", Emoji: "🌀", Weight: 120},                     // 12% (Реролл без кулдауна)
	{ID: 9, Name: "Бесплатная доставка", Emoji: "🛵", Weight: 50},         // 5%  (Промокод на 360 руб)
}

//easyjson:json
type WheelSpinResponse struct {
	SectorID   int     `json:"sector_id"`
	SectorName string  `json:"sector_name"`
	Emoji      string  `json:"emoji"`
	PromoCode  *string `json:"promo_code,omitempty"`
	ExpiresAt  *string `json:"expires_at,omitempty"`
	Message    string  `json:"message"`
}

type WheelHandler struct {
	userClient  userclient.UserClient
	orderClient orderclient.OrderClient
	promoClient pbOrder.PromoServiceClient
	logger      logger.Logger
}

func NewWheelHandler(
	uc userclient.UserClient,
	oc orderclient.OrderClient,
	pc pbOrder.PromoServiceClient,
	l logger.Logger,
) *WheelHandler {
	return &WheelHandler{
		userClient:  uc,
		orderClient: oc,
		promoClient: pc,
		logger:      l,
	}
}

// Spin godoc
// @Summary 		Запуск Колеса Фортуны (Lucky Wheel)
// @Description		Проверяет кулдаун, разыгрывает сектор с призами, начисляет награду. Кулдаун хранится в БД user_service.
// @Tags			profile
// @Produce			json
// @Success			200		{object}  WheelSpinResponse
// @Failure			400		{object}  response.ErrorResponse "Кулдаун активен или неверный запрос"
// @Failure			401		{object}  response.ErrorResponse "Неавторизован"
// @Failure			500		{object}  response.ErrorResponse "Внутренняя ошибка"
// @Router			/profile/wheel/spin [post]
func (h *WheelHandler) Spin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err := h.userClient.ClaimWheelSpin(ctx, userID)
	if err != nil {
		if err == userclient.ErrWheelCooldownActive {
			response.Error(w, http.StatusBadRequest, "Вы уже крутили колесо сегодня. Попробуйте завтра!")
			return
		}
		if err == userclient.ErrUserNotFound {
			response.Error(w, http.StatusNotFound, "Пользователь не найден")
			return
		}
		l.Error("failed to claim wheel spin via gRPC", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Разыгрываем случайный сектор на основе весов
	totalWeight := 0
	for _, s := range sectors {
		totalWeight += s.Weight
	}

	randVal := rand.Intn(totalWeight)
	var wonSector Sector
	currentSum := 0

	for _, s := range sectors {
		currentSum += s.Weight
		if randVal < currentSum {
			wonSector = s
			break
		}
	}

	resp := WheelSpinResponse{
		SectorID:   wonSector.ID,
		SectorName: wonSector.Name,
		Emoji:      wonSector.Emoji,
	}

	switch wonSector.ID {
	case 1: // Попробуй в следующий раз
		resp.Message = "Увы, в этот раз ничего не выпало. Попробуйте завтра!"

	case 2: // Персональная скидка (150 руб)
		discountAmount := int64(150_000_000)
		promo, err := h.promoClient.CreateAndBindWheelPromo(ctx, &pbOrder.CreateAndBindWheelPromoRequest{
			UserId:         userID,
			DiscountAmount: &discountAmount,
			Title:          "Персональная скидка 150 рублей",
		})
		if err != nil {
			l.Error("failed to generate wheel promo (150 rub)", err)
			response.Error(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		resp.PromoCode = &promo.Code
		resp.ExpiresAt = promo.ExpiresAt
		resp.Message = "Поздравляем! Вы выиграли персональную скидку 150 рублей на любой заказ!"

	case 3: // Скидка на популярный бренд (15%)
		// Получаем популярные бренды за последние 7 дней
		brandIDs, err := h.orderClient.GetTrendingBrands(ctx, 7, 10)
		var selectedBrandID int64
		if err != nil || len(brandIDs) == 0 {
			selectedBrandID = 1
		} else {
			selectedBrandID = brandIDs[rand.Intn(len(brandIDs))]
		}

		discountPercent := int32(15)
		promo, err := h.promoClient.CreateAndBindWheelPromo(ctx, &pbOrder.CreateAndBindWheelPromoRequest{
			UserId:            userID,
			DiscountPercent:   &discountPercent,
			RestaurantBrandId: &selectedBrandID,
			Title:             "Скидка 15% на популярный бренд",
		})
		if err != nil {
			l.Error("failed to generate wheel promo (15%)", err)
			response.Error(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		resp.PromoCode = &promo.Code
		resp.ExpiresAt = promo.ExpiresAt
		resp.Message = "Вы выиграли скидку 15% на популярный ресторан!"

	case 4: // Заморозка стрика
		err := h.userClient.ActivateStreakFreeze(ctx, userID)
		if err != nil {
			l.Error("failed to activate streak freeze via wheel", err)
			response.Error(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		resp.Message = "Заморозка серии активна! Теперь ваша серия заказов защищена, если вы пропустите неделю."

	case 5: // Стрик-буст (+1 неделя)
		err := h.userClient.IncrementStreak(ctx, userID)
		if err != nil {
			l.Error("failed to increment streak via wheel", err)
			response.Error(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		resp.Message = "Ура! Ваша серия заказов увеличена на +1 неделю!"

	case 6: // Эксклюзивная ачивка (Любимчик Пиццули)
		err := h.userClient.OnWheelSpin(ctx, userID, "lucky_wheel_winner")
		if err != nil {
			l.Error("failed to award exclusive wheel achievement", err)
			response.Error(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		resp.Message = "Невероятное везение! Вы получили редкую коллекционную ачивку «Любимчик Пиццули»!"

	case 7: // Супер-приз (500 рублей от 1500 рублей)
		discountAmount := int64(500_000_000)
		minOrderAmount := int64(1_500_000_000)
		promo, err := h.promoClient.CreateAndBindWheelPromo(ctx, &pbOrder.CreateAndBindWheelPromoRequest{
			UserId:         userID,
			DiscountAmount: &discountAmount,
			MinOrderAmount: &minOrderAmount,
			Title:          "Супер-приз: скидка 500 рублей",
		})
		if err != nil {
			l.Error("failed to generate wheel super promo (500 rub)", err)
			response.Error(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		resp.PromoCode = &promo.Code
		resp.ExpiresAt = promo.ExpiresAt
		resp.Message = "Мега-приз! Вы выиграли скидку 500 рублей на заказ от 1500 рублей!"

	case 8: // Еще одна попытка (Реролл)
		err := h.userClient.ResetWheelSpinCooldown(ctx, userID)
		if err != nil {
			l.Error("failed to reset wheel cooldown in user_service", err)
			response.Error(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		resp.Message = "Вам выпала еще одна попытка! Вращайте колесо прямо сейчас!"

	case 9: // Бесплатная доставка (Промокод на 360 руб)
		discountAmount := int64(360_000_000)
		promo, err := h.promoClient.CreateAndBindWheelPromo(ctx, &pbOrder.CreateAndBindWheelPromoRequest{
			UserId:         userID,
			DiscountAmount: &discountAmount,
			Title:          "Бесплатная доставка (скидка 360 рублей)",
		})
		if err != nil {
			l.Error("failed to generate wheel free delivery promo", err)
			response.Error(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		resp.PromoCode = &promo.Code
		resp.ExpiresAt = promo.ExpiresAt
		resp.Message = "Поздравляем! Вы выиграли промокод на бесплатную доставку!"
	}

	// Если выиграна не коллекционная ачивка, вызываем OnWheelSpin без кода ачивки
	if wonSector.ID != 6 {
		_ = h.userClient.OnWheelSpin(ctx, userID, "")
	}

	l.Info("user successfully spun lucky wheel",
		logger.Int64("user_id", userID),
		logger.Int("sector_id", wonSector.ID),
		logger.String("sector_name", wonSector.Name),
	)

	response.JSON(w, http.StatusOK, resp)
}

// GetSectors godoc
// @Summary 		Получение списка секторов Колеса Фортуны
// @Description		Возвращает список всех секторов (ID, имя, эмодзи) для динамической отрисовки на фронтенде. Веса (вероятности) скрыты.
// @Tags			profile
// @Produce			json
// @Success			200		{object}  WheelSectorsResponse
// @Failure			401		{object}  response.ErrorResponse "Неавторизован"
// @Router			/profile/wheel/sectors [get]
func (h *WheelHandler) GetSectors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	dtoSectors := make([]SectorResponse, 0, len(sectors))
	for _, s := range sectors {
		dtoSectors = append(dtoSectors, SectorResponse{
			ID:    s.ID,
			Name:  s.Name,
			Emoji: s.Emoji,
		})
	}

	response.JSON(w, http.StatusOK, WheelSectorsResponse{Sectors: dtoSectors})
}
