package events

//go:generate easyjson $GOFILE

const (
	QueueAnalytics = "queue.analytics.clickhouse"
)

// Структура для передачи заказа, соответствует orders_report_log
//
//easyjson:json
type AnalyticsOrderEvent struct {
	EventTime int64 `json:"event_time"`

	OrderPublicID string `json:"order_public_id"`
	RestaurantID  int64  `json:"restaurant_id"`
	ClientID      int64  `json:"client_id"` // Создатель корзины

	TotalCostRaw         int64 `json:"total_cost_raw"`
	DeliveryCostRaw      int64 `json:"delivery_cost_raw"`
	ServiceFeeRaw        int64 `json:"service_fee_raw"`
	DiscountRaw          int64 `json:"discount_raw"`
	RestaurantRevenueRaw int64 `json:"restaurant_revenue_raw"`

	Status     string `json:"status"`
	PrevStatus string `json:"prev_status"`

	OrderType    string `json:"order_type"`    // solo или shared
	MembersCount int32  `json:"members_count"` // Кол-во человек в корзине
	City         string `json:"city"`

	// Должен быть 1 только если статус заказа перешел в paid
	IsFinancialImpact int8 `json:"is_financial_impact"`

	// Список блюд для таблицы order_items_report_log
	// Заполняется только при статусе paid
	Items []AnalyticsOrderItem `json:"items,omitempty"`
}

// Структура для передачи блюд, соответствует order_items_report_log
//
//easyjson:json
type AnalyticsOrderItem struct {
	DishID   int64  `json:"dish_id"`
	DishName string `json:"dish_name"`

	Quantity    int32 `json:"quantity"`
	PriceRaw    int64 `json:"price_raw"`
	RowTotalRaw int64 `json:"row_total_raw"`

	// Кому в совместной корзине принадлежит блюдо
	UserID int64 `json:"user_id"`
}
