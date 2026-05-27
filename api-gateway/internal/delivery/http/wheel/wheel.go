package wheel

//go:generate easyjson $GOFILE

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/userclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/response"
	"google.golang.org/grpc/status"
)

// Описание секторов колеса (нужно для метода GetSectors)
type Sector struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Emoji string `json:"emoji"`
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

type WheelHandler struct {
	userClient userclient.UserClient // <-- Оставили только связь с user-service
	logger     logger.Logger
}

// Упростили конструктор, убрав 2 лишние зависимости
func NewWheelHandler(
	uc userclient.UserClient,
	l logger.Logger,
) *WheelHandler {
	return &WheelHandler{
		userClient: uc,
		logger:     l,
	}
}

// Spin godoc
// @Summary 		Запуск Колеса Фортуны (Lucky Wheel)
// @Description		Делегирует проверку кулдауна и розыгрыш приза микросервису пользователей
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

	resp, err := h.userClient.SpinWheel(ctx, userID)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Message() == "WHEEL_COOLDOWN_ACTIVE" {
				response.Error(w, http.StatusBadRequest, "Вы уже крутили колесо сегодня. Попробуйте завтра!")
				return
			}
		}
		l.Error("lucky wheel system failure", err,
			logger.Int64("user_id", userID),
		)

		response.Error(w, http.StatusInternalServerError, "К сожалению, колесо Пиццули не крутится! Мы уже чиним, попробуйте позже!")
		return
	}

	l.Debug("user successfully spun lucky wheel via microservice",
		logger.Int64("user_id", userID),
		logger.String("sector_name", resp.SectorName),
	)

	spinResult := WheelSpinResponse{
		SectorID:   int(resp.SectorId),
		SectorName: resp.SectorName,
		Emoji:      resp.Emoji,
		Message:    resp.Message,
	}
	if resp.PromoCode != nil {
		spinResult.PromoCode = resp.PromoCode
	}
	if resp.ExpiresAt != nil {
		spinResult.ExpiresAt = resp.ExpiresAt
	}

	response.JSON(w, http.StatusOK, spinResult)
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

	dbSectors, err := h.userClient.GetWheelSectors(ctx)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	dtoSectors := make([]SectorResponse, 0, len(dbSectors))
	for _, s := range dbSectors {
		dtoSectors = append(dtoSectors, SectorResponse{
			ID:    s.ID,
			Name:  s.Name,
			Emoji: s.Emoji,
		})
	}

	response.JSON(w, http.StatusOK, WheelSectorsResponse{Sectors: dtoSectors})
}
