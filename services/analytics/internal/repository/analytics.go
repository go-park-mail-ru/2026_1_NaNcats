package repository

//go:generate mockgen -destination=mocks/analytics_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/repository AnalyticsRepository

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"
)

// FinancialStats описывает финансовые показатели ресторана за период.
// Все денежные значения представлены в сырых единицах с точностью до 6 знаков после запятой (1 рубль = 1 000 000)
type FinancialStats struct {
	TotalRevenueRaw   int64 // Общая выручка (чистая, без доставки)
	AverageTicketRaw  int64 // Средний чек
	TotalDiscountsRaw int64 // Общая сумма предоставленных скидок по промокодам
	TotalOrdersCount  int64 // Общее количество созданных заказов за период
}

// Описывает операционную эффективность кухни и воронку заказов
type OperationalStats struct {
	AvgCookingTimeSec int64            // Среднее время готовки на кухне в секундах
	StatusCounts      map[string]int64 // Количество заказов для каждого статуса
}

// Описывает блюдо-лидер продаж
type BestSeller struct {
	DishID          int64
	DishName        string
	UnitsSold       int32 // Общее количество проданных порций
	TotalRevenueRaw int64 // Общая выручка от продаж этого блюда
}

// Статистика по формату заказа (соло/совместный)
type OrderTypeStat struct {
	OrderType    string  // "solo" или "shared"
	OrdersCount  int64   // Количество заказов этого типа
	AvgGroupSize float64 // Среднее количество участников заказа
}

// Представляет одну временнУю точку для построения графиков на фронтенде
type DailyStat struct {
	Date        time.Time // День, к которому относится срез
	RevenueRaw  int64     // Выручка за этот день
	OrdersCount int64     // Количество заказов за этот день
}

// Все аналитические данные дашборда владельца в единой структуре
type OwnerStats struct {
	Financial   FinancialStats
	Operational OperationalStats
	Dishes      []BestSeller
	OrderTypes  []OrderTypeStat
	Timeline    []DailyStat
}

type AnalyticsRepository interface {
	// Ставит событие в очередь для последующей пакетной вставки
	InsertEvent(ctx context.Context, event events.AnalyticsOrderEvent) error

	// Извлекает статистику по ресторану за выбранный период времени из ClickHouse СУБД.
	GetOwnerStats(ctx context.Context, restaurantID int64, startTime, endTime time.Time) (OwnerStats, error)

	// Инициирует контролируемое завершение работы СУБД-адаптера
	Close() error
}
