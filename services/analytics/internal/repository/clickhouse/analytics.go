package clickhouse

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	rabbitmqErrors "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/errors"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"
	"github.com/google/uuid"
)

// Событие и уже распарсенный UUID заказа
type bufferedEvent struct {
	event     events.AnalyticsOrderEvent
	orderUUID uuid.UUID
}

type clickhouseRepo struct {
	conn      driver.Conn
	logger    logger.Logger
	eventChan chan bufferedEvent
	stopChan  chan struct{}
	wg        sync.WaitGroup // Отслеживает старт/стоп самого воркера

	// Поля для безопасного graceful shutdown
	mu        sync.RWMutex   // Защищает флаг closed и канал eventChan
	writersWg sync.WaitGroup // Отслеживает активные записи в метод InsertEvent
	closed    bool           // Флаг закрытия репозитория
}

func NewAnalyticsRepository(conn driver.Conn, l logger.Logger, batchSize int, flushInterval time.Duration) repository.AnalyticsRepository {
	repo := &clickhouseRepo{
		conn:      conn,
		logger:    l,
		eventChan: make(chan bufferedEvent, batchSize*2),
		stopChan:  make(chan struct{}),
	}

	repo.wg.Add(1)
	go repo.startWorker(batchSize, flushInterval)

	return repo
}

func (r *clickhouseRepo) Close() error {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return nil
	}
	r.closed = true
	close(r.stopChan) // Разблокирует потоки, застрявшие в select на отправке
	r.mu.Unlock()

	// Ждем, пока все активные вызовы InsertEvent завершат свою работу
	r.writersWg.Wait()

	// Теперь больше никто не пишет и не попытается написать в eventChan, закрываем канал
	close(r.eventChan)

	// Ждем, пока воркер полностью вычитает все оставшиеся в буфере канала сообщения и завершится
	r.wg.Wait()

	return nil
}

func (r *clickhouseRepo) InsertEvent(ctx context.Context, event events.AnalyticsOrderEvent) error {
	r.mu.RLock()
	if r.closed {
		r.mu.RUnlock()
		return fmt.Errorf("repository is closed, rejecting new events")
	}

	r.writersWg.Add(1)
	r.mu.RUnlock()

	defer r.writersWg.Done()

	orderUUID, err := uuid.Parse(event.OrderPublicID)
	if err != nil {
		return rabbitmqErrors.NewPermanentError(fmt.Errorf("invalid order public id format [%s]: %w", event.OrderPublicID, err))
	}

	be := bufferedEvent{
		event:     event,
		orderUUID: orderUUID,
	}

	select {
	case r.eventChan <- be:
		return nil
	case <-r.stopChan:
		return fmt.Errorf("repository was closed during insert")
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *clickhouseRepo) startWorker(batchSize int, flushInterval time.Duration) {
	defer r.wg.Done()

	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	buffer := make([]bufferedEvent, 0, batchSize)

	for {
		select {
		// Читаем из канала и отслеживаем его закрытие
		case be, ok := <-r.eventChan:
			if !ok {
				// Канал закрыт и полностью вычитан, сбрасываем остатки и выходим
				if len(buffer) > 0 {
					r.flush(buffer)
				}
				return
			}
			buffer = append(buffer, be)
			if len(buffer) >= batchSize {
				r.flush(buffer)
				buffer = buffer[:0]
			}

		case <-ticker.C:
			if len(buffer) > 0 {
				r.flush(buffer)
				buffer = buffer[:0]
			}
		}
	}
}

func (r *clickhouseRepo) flush(batch []bufferedEvent) {
	if len(batch) == 0 {
		return
	}

	writeCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	orderBatch, err := r.conn.PrepareBatch(writeCtx, `
		INSERT INTO orders_report_log (
			event_time, order_public_id, restaurant_id, client_id,
			total_cost_raw, delivery_cost_raw, service_fee_raw, discount_raw,
			restaurant_revenue_raw, status, prev_status, order_type,
			members_count, city, is_financial_impact
		)
	`)
	if err != nil {
		r.logger.Error("failed to prepare orders batch for ClickHouse", err)
		return
	}

	var hasItems bool
	for _, be := range batch {
		t := time.UnixMilli(be.event.EventTime)
		err = orderBatch.Append(
			t, be.orderUUID, be.event.RestaurantID, be.event.ClientID,
			be.event.TotalCostRaw, be.event.DeliveryCostRaw, be.event.ServiceFeeRaw, be.event.DiscountRaw,
			be.event.RestaurantRevenueRaw, be.event.Status, be.event.PrevStatus, be.event.OrderType,
			be.event.MembersCount, be.event.City, be.event.IsFinancialImpact,
		)
		if err != nil {
			r.logger.Error("failed to append to orders batch", err)
		}
		if len(be.event.Items) > 0 {
			hasItems = true
		}
	}

	if err := orderBatch.Send(); err != nil {
		r.logger.Error("failed to send orders batch to ClickHouse", err)
		return
	}

	if hasItems {
		itemBatch, err := r.conn.PrepareBatch(writeCtx, `
			INSERT INTO order_items_report_log (
				event_time, order_public_id, restaurant_id, dish_id,
				dish_name, quantity, price_raw, row_total_raw, user_id
			)
		`)
		if err != nil {
			r.logger.Error("failed to prepare items batch for ClickHouse", err)
			return
		}

		for _, be := range batch {
			t := time.UnixMilli(be.event.EventTime)
			for _, item := range be.event.Items {
				err = itemBatch.Append(
					t, be.orderUUID, be.event.RestaurantID, item.DishID,
					item.DishName, item.Quantity, item.PriceRaw, item.RowTotalRaw, item.UserID,
				)
				if err != nil {
					r.logger.Error("failed to append to items batch", err)
				}
			}
		}

		if err := itemBatch.Send(); err != nil {
			r.logger.Error("failed to send items batch to ClickHouse", err)
		}
	}

	r.logger.Debug("successfully flushed analytics batch to ClickHouse", logger.Int("batch_size", len(batch)))
}

func (r *clickhouseRepo) GetOwnerStats(ctx context.Context, restaurantID int64, startTime, endTime time.Time) (repository.OwnerStats, error) {
	stats := repository.OwnerStats{
		Operational: repository.OperationalStats{
			StatusCounts: make(map[string]int64),
		},
		Dishes:     make([]repository.BestSeller, 0),
		OrderTypes: make([]repository.OrderTypeStat, 0),
		Timeline:   make([]repository.DailyStat, 0),
	}

	// Финансовые показатели
	financialQuery := `
		SELECT 
			toInt64(uniqExact(order_public_id)) AS total_orders,
			toInt64(uniqExactIf(order_public_id, is_financial_impact = 1)) AS paid_orders,
			toInt64(sumIf(restaurant_revenue_raw, is_financial_impact = 1)) AS net_revenue,
			toInt64(sumIf(discount_raw, is_financial_impact = 1)) AS total_discounts
		FROM orders_report_log
		WHERE restaurant_id = ?
		  AND event_time >= ?
		  AND event_time <= ?;
	`

	var totalOrders int64
	var paidOrders int64
	var netRevenue int64
	var totalDiscounts int64

	err := r.conn.QueryRow(ctx, financialQuery, restaurantID, startTime, endTime).Scan(&totalOrders, &paidOrders, &netRevenue, &totalDiscounts)
	if err != nil {
		return stats, fmt.Errorf("query financial stats failed: %w", err)
	}

	stats.Financial.TotalRevenueRaw = netRevenue
	stats.Financial.TotalDiscountsRaw = totalDiscounts
	stats.Financial.TotalOrdersCount = totalOrders

	if paidOrders > 0 {
		stats.Financial.AverageTicketRaw = netRevenue / paidOrders
	}

	// Среднее время готовки
	cookingTimeQuery := `
		SELECT 
			toInt64(nanToZero(avg(waiting_time - progress_time))) AS avg_cooking_time_sec
		FROM (
			SELECT 
				order_public_id,
				maxIf(toUnixTimestamp(event_time), status = 'waiting') AS waiting_time,
				maxIf(toUnixTimestamp(event_time), status = 'in_progress') AS progress_time
			FROM orders_report_log
			WHERE restaurant_id = ?
			  AND event_time >= ?
			  AND event_time <= ?
			GROUP BY order_public_id
			HAVING waiting_time > 0 AND progress_time > 0
		);
	`

	err = r.conn.QueryRow(ctx, cookingTimeQuery, restaurantID, startTime, endTime).Scan(&stats.Operational.AvgCookingTimeSec)
	if err != nil {
		return stats, fmt.Errorf("query cooking time failed: %w", err)
	}

	// Уникальное распределение текущих статусов заказов за период
	statusFunnelQuery := `
		SELECT status, toInt64(count(*)) AS count
		FROM (
			SELECT 
				order_public_id,
				argMax(status, event_time) AS status
			FROM orders_report_log
			WHERE restaurant_id = ?
			  AND event_time >= ?
			  AND event_time <= ?
			GROUP BY order_public_id
		)
		GROUP BY status;
	`

	statusRows, err := r.conn.Query(ctx, statusFunnelQuery, restaurantID, startTime, endTime)
	if err != nil {
		return stats, fmt.Errorf("query status funnel failed: %w", err)
	}
	defer statusRows.Close()

	for statusRows.Next() {
		var status string
		var count int64
		if err := statusRows.Scan(&status, &count); err != nil {
			return stats, fmt.Errorf("scan status funnel row failed: %w", err)
		}
		stats.Operational.StatusCounts[status] = count
	}

	// Топ-10 продаваемых блюд
	bestSellersQuery := `
		SELECT 
			dish_id,
			any(dish_name) AS name,
			toInt32(COALESCE(sum(quantity), 0)) AS units_sold,
			toInt64(COALESCE(sum(row_total_raw), 0)) AS total_revenue
		FROM order_items_report_log
		WHERE restaurant_id = ?
		  AND event_time >= ?
		  AND event_time <= ?
		GROUP BY dish_id
		ORDER BY units_sold DESC
		LIMIT 10;
	`

	dishRows, err := r.conn.Query(ctx, bestSellersQuery, restaurantID, startTime, endTime)
	if err != nil {
		return stats, fmt.Errorf("query best sellers failed: %w", err)
	}
	defer dishRows.Close()

	for dishRows.Next() {
		var dish repository.BestSeller
		err := dishRows.Scan(&dish.DishID, &dish.DishName, &dish.UnitsSold, &dish.TotalRevenueRaw)
		if err != nil {
			return stats, fmt.Errorf("scan best seller row failed: %w", err)
		}
		stats.Dishes = append(stats.Dishes, dish)
	}

	// Разделение типов заказов (Solo vs Shared)
	orderTypesQuery := `
		SELECT 
			order_type,
			toInt64(count(DISTINCT order_public_id)) AS orders_count,
			if(isNaN(avg(members_count)), 1.0, avg(members_count)) AS avg_group_size
		FROM orders_report_log
		WHERE restaurant_id = ?
		  AND event_time >= ?
		  AND event_time <= ?
		  AND status = 'cart_locked' 
		GROUP BY order_type;
	`

	typeRows, err := r.conn.Query(ctx, orderTypesQuery, restaurantID, startTime, endTime)
	if err != nil {
		return stats, fmt.Errorf("query order types failed: %w", err)
	}
	defer typeRows.Close()

	for typeRows.Next() {
		var stat repository.OrderTypeStat
		err := typeRows.Scan(&stat.OrderType, &stat.OrdersCount, &stat.AvgGroupSize)
		if err != nil {
			return stats, fmt.Errorf("scan order type row failed: %w", err)
		}
		stats.OrderTypes = append(stats.OrderTypes, stat)
	}

	// ВременнАя шкала по дням
	timelineQuery := `
		SELECT 
			toStartOfDay(event_time) AS day,
			COALESCE(sum(restaurant_revenue_raw) FILTER (WHERE is_financial_impact = 1), 0) AS daily_revenue,
			toInt64(count(DISTINCT order_public_id)) AS daily_orders
		FROM orders_report_log
		WHERE restaurant_id = ?
		  AND event_time >= ?
		  AND event_time <= ?
		GROUP BY day
		ORDER BY day ASC;
	`

	timelineRows, err := r.conn.Query(ctx, timelineQuery, restaurantID, startTime, endTime)
	if err != nil {
		return stats, fmt.Errorf("query timeline failed: %w", err)
	}
	defer timelineRows.Close()

	for timelineRows.Next() {
		var pt repository.DailyStat
		err := timelineRows.Scan(&pt.Date, &pt.RevenueRaw, &pt.OrdersCount)
		if err != nil {
			return stats, fmt.Errorf("scan timeline row failed: %w", err)
		}
		stats.Timeline = append(stats.Timeline, pt)
	}

	return stats, nil
}
