package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"
	"github.com/google/uuid"
)

type clickhouseRepo struct {
	conn driver.Conn
}

func NewAnalyticsRepository(conn driver.Conn) repository.AnalyticsRepository {
	return &clickhouseRepo{conn: conn}
}

func (r *clickhouseRepo) InsertBatch(ctx context.Context, data []events.AnalyticsOrderEvent) error {
	orderBatch, err := r.conn.PrepareBatch(ctx, `
		INSERT INTO orders_report_log (
			event_time, order_public_id, restaurant_id, client_id,
			total_cost_raw, delivery_cost_raw, service_fee_raw, discount_raw,
			restaurant_revenue_raw, status, prev_status, order_type,
			members_count, city, is_financial_impact
		)
	`)
	if err != nil {
		return fmt.Errorf("prepare order batch: %w", err)
	}

	itemBatch, err := r.conn.PrepareBatch(ctx, `
		INSERT INTO order_items_report_log (
			event_time, order_public_id, restaurant_id, dish_id,
			dish_name, quantity, price_raw, row_total_raw, user_id
		)
	`)
	if err != nil {
		return fmt.Errorf("prepare items batch: %w", err)
	}

	for _, event := range data {
		orderUUID, _ := uuid.Parse(event.OrderPublicID)
		t := time.UnixMilli(event.EventTime)

		err = orderBatch.Append(
			t, orderUUID, event.RestaurantID, event.ClientID,
			event.TotalCostRaw, event.DeliveryCostRaw, event.ServiceFeeRaw, event.DiscountRaw,
			event.RestaurantRevenueRaw, event.Status, event.PrevStatus, event.OrderType,
			event.MembersCount, event.City, event.IsFinancialImpact,
		)
		if err != nil {
			return err
		}

		for _, item := range event.Items {
			err = itemBatch.Append(
				t, orderUUID, event.RestaurantID, item.DishID,
				item.DishName, item.Quantity, item.PriceRaw, item.RowTotalRaw, item.UserID,
			)
			if err != nil {
				return err
			}
		}
	}

	if err := orderBatch.Send(); err != nil {
		return err
	}
	if itemBatch.Rows() > 0 {
		return itemBatch.Send()
	}

	return nil
}
