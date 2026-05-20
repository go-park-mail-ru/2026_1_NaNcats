package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"
)

type AnalyticsRepository interface {
	InsertEvent(ctx context.Context, event events.AnalyticsOrderEvent) error
}
