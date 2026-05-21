package clickhouse

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
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
	stopChan  chan struct{} // Сигнальный канал для остановки воркера
	wg        sync.WaitGroup
	closed    atomic.Bool // Для безопасной проверки состояния
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
	// Атомарно меняем флаг. Если уже закрыт - выходим
	if r.closed.Swap(true) {
		return nil
	}

	// Закрываем сигнальный канал. Это разблокирует все горутины, слушающие r.stopChan в select
	close(r.stopChan)

	// Ждем, пока воркер обработает остатки из eventChan и запишет их в CH
	r.wg.Wait()

	return nil
}

func (r *clickhouseRepo) InsertEvent(ctx context.Context, event events.AnalyticsOrderEvent) error {
	// Атомарная быстрая проверка
	if r.closed.Load() {
		return fmt.Errorf("repository is closed, rejecting new events")
	}

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
		case be := <-r.eventChan:
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

		case <-r.stopChan:
			// Graceful Shutdown
			// Неблокирующе читаем все сообщения которые успели накопиться в eventChan до закрытия приложения и сливаем их
			for {
				select {
				case be := <-r.eventChan:
					buffer = append(buffer, be)
					if len(buffer) >= batchSize {
						r.flush(buffer)
						buffer = buffer[:0]
					}
				default:
					// Канал пуст
					if len(buffer) > 0 {
						r.flush(buffer)
					}
					return
				}
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
