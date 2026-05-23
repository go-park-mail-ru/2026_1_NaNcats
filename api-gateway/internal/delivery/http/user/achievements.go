package user

//go:generate easyjson $GOFILE

import (
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/response"
)

//easyjson:json
type AchievementItem struct {
	Code        string `json:"code"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Earned      bool   `json:"earned"`
	AwardedAt   string `json:"awarded_at,omitempty"`
}

//easyjson:json
type AchievementsResponse struct {
	Items []AchievementItem `json:"items"`
}

// GetAchievements godoc
// @Summary 		Список ачивок пользователя
// @Description		Все ачивки системы с флагом earned для текущего пользователя
// @Tags			profile
// @Produce			json
// @Success			200	{object}  AchievementsResponse
// @Failure			500	{object}  response.ErrorResponse
// @Router			/profile/achievements [get]
func (h *UserProfileHandler) GetAchievements(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	all, err := h.userClient.ListAchievements(ctx)
	if err != nil {
		l.Error("failed to list achievements", err)
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	earned, err := h.userClient.GetUserAchievements(ctx, userID)
	if err != nil {
		l.Error("failed to get user achievements", err)
		response.Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	earnedMap := make(map[int64]time.Time, len(earned))
	for _, e := range earned {
		earnedMap[e.AchievementID] = e.AwardedAt
	}

	items := make([]AchievementItem, 0, len(all))
	for _, a := range all {
		item := AchievementItem{
			Code:        a.Code,
			Title:       a.Title,
			Description: a.Description,
			Icon:        a.Icon,
		}
		if at, ok := earnedMap[a.ID]; ok {
			item.Earned = true
			item.AwardedAt = at.UTC().Format(time.RFC3339)
		}
		items = append(items, item)
	}

	response.JSON(w, http.StatusOK, AchievementsResponse{Items: items})
}
