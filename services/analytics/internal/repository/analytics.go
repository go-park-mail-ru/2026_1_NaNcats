package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"
)

type AnalyticsRepository interface {
	InsertBatch(ctx context.Context, data []events.AnalyticsOrderEvent) error
}
