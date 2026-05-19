package usecase

import (
	"context"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"
)

type AnalyticsUseCase interface {
	// Принимает событие из RabbitMQ и кладет его в буфер
	Collect(ctx context.Context, event events.AnalyticsOrderEvent) error
	// Запускает фоновый воркер сброса данных по времени и корректной обработки graceful shutdown
	Start(ctx context.Context)
}

type analyticsUseCase struct {
	conn          driver.Conn
	batchSize     int
	flushInterval time.Duration
	logger        logger.Logger

	// Буфер для накопления событий
	mu     sync.Mutex
	buffer []events.AnalyticsOrderEvent
}

func NewAnalyticsUseCase(conn driver.Conn, batchSize int, interval time.Duration, l logger.Logger) AnalyticsUseCase {
	return &analyticsUseCase{
		conn:          conn,
		batchSize:     batchSize,
		flushInterval: interval,
		logger:        l,
		buffer:        make([]events.AnalyticsOrderEvent, 0, batchSize),
	}
}

func (u *analyticsUseCase) flush(ctx context.Context) {
	u.mu.Lock()
	if len(u.buffer) == 0 {
		u.mu.Unlock()
		return
	}

	batchData := u.buffer
	u.buffer = make([]events.AnalyticsOrderEvent, 0, u.batchSize)
	u.mu.Unlock()

	u.logger.Debug("flushing batch to clickhouse", logger.Int("count", len(batchData)))

	if err := u.writeToClickHouse(ctx, batchData); err != nil {
		u.logger.Error("failed to write batch to clickhouse", err)
		// TODO: вернуть данные в буфер или сохранить в файл
	}
}

func (u *analyticsUseCase) Start(ctx context.Context) {
	ticker := time.NewTicker(u.flushInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				u.flush(context.Background())
			case <-ctx.Done():
				ticker.Stop()
				// Финальный сброс перед выключением
				u.flush(context.Background())
				return
			}
		}
	}()
}

func (u *analyticsUseCase) Collect(ctx context.Context, event events.AnalyticsOrderEvent) error {
	u.mu.Lock()
	u.buffer = append(u.buffer, event)
	readyToFlush := len(u.buffer) >= u.batchSize
	u.mu.Unlock()

	// Если набрали достаточно данных сбрасываем в ClickHouse
	if readyToFlush {
		go u.flush(context.Background())
	}

	return nil
}

func (u *analyticsUseCase) writeToClickHouse(ctx context.Context, data []events.AnalyticsOrderEvent) error {
	// Временно заглушка
	return nil
}
