package analytics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/analyticsclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/response"
)

type AnalyticsHandler struct {
	analyticsClient analyticsclient.AnalyticsClient
	logger          logger.Logger
}

func NewAnalyticsHandler(ac analyticsclient.AnalyticsClient, l logger.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsClient: ac,
		logger:          l,
	}
}

// GetOwnerAnalytics godoc
// @Summary 		Получение аналитики дашборда (для владельцев)
// @Description		Возвращает финансовую, операционную и товарную статистику ресторана с ролевой защитой.
// @Tags			owner, analytics
// @Produce			json
// @Param			restaurant_id	query	  int		true	"ID ресторана"
// @Param			start_time		query	  string	true	"Начало периода (ГГГГ-ММ-ДД)"
// @Param			end_time		query	  string	true	"Конец периода (ГГГГ-ММ-ДД)"
// @Success			200				{object}  analyticsclient.OwnerStats
// @Failure			401				{object}  response.ErrorResponse "Неавторизован"
// @Failure			403				{object}  response.ErrorResponse "Отказ в доступе (не ваш ресторан)"
// @Failure			404				{object}  response.ErrorResponse "Данные не найдены"
func (h *AnalyticsHandler) GetOwnerAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := h.logger.WithContext(ctx)

	restaurantIDStr := r.URL.Query().Get("restaurant_id")
	restaurantID, err := strconv.ParseInt(restaurantIDStr, 10, 64)
	if err != nil || restaurantID <= 0 {
		response.Error(w, http.StatusBadRequest, "invalid or missing restaurant_id query parameter")
		return
	}

	// Читаем и парсим временнЫе рамки в формате YYYY-MM-DD
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")

	if startTimeStr == "" || endTimeStr == "" {
		response.Error(w, http.StatusBadRequest, "start_time and end_time query parameters are required")
		return
	}

	startTime, err := time.Parse("2006-01-02", startTimeStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid start_time format (expected YYYY-MM-DD)")
		return
	}

	endTime, err := time.Parse("2006-01-02", endTimeStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid end_time format (expected YYYY-MM-DD)")
		return
	}
	// Сдвигаем endTime на конец дня, иначе ClickHouse-фильтр event_time <= '2026-05-27 00:00:00'
	// отрезает все сегодняшние заказы.
	endTime = endTime.Add(24*time.Hour - time.Nanosecond)

	stats, err := h.analyticsClient.GetOwnerStats(ctx, restaurantID, startTime, endTime)
	if err != nil {
		l.Error("failed to get restaurant owner stats over grpc", err, logger.Int64("restaurant_id", restaurantID))
		response.WriteError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, stats)
}
