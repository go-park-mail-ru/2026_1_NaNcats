// Package autoadvance - фоновый симулятор кухни/курьера
// Каждый тик берёт неотложные заказы из БД и продвигает их вперёд по
// фиксированной цепочке: paid -> in_progress -> delivering -> finished.
// После UPDATE публикует GatewayEvent в RabbitMQ - фронт получает live-апдейт
// через WebSocket
package autoadvance

import (
	"context"
	"sort"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"
)

func sourceStatuses() []string {
	out := make([]string, 0, len(transitions))
	for k := range transitions {
		out = append(out, k)
	}
	sort.Strings(out)

	return out
}

type Publisher interface {
	PublishJSON(ctx context.Context, queue string, payload any) error
}

var transitions = map[string]string{
	"paid":        "in_progress",
	"in_progress": "delivering",
	"waiting":     "delivering",
	"delivering":  "finished",
}

type Runner struct {
	repo      repository.OrderRepository
	publisher Publisher
	interval  time.Duration
	logger    logger.Logger
}

func New(repo repository.OrderRepository, pub Publisher, interval time.Duration, l logger.Logger) *Runner {
	return &Runner{
		repo:      repo,
		publisher: pub,
		interval:  interval,
		logger:    l,
	}
}

func (r *Runner) Run(ctx context.Context) {
	r.logger.Info("Order auto-advancer started", logger.String("interval", r.interval.String()))
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("Order auto-advancer stopped")
			return
		case <-ticker.C:
			r.tick(ctx)
		}
	}
}

func (r *Runner) tick(ctx context.Context) {
	orders, err := r.repo.GetOrdersByStatuses(ctx, sourceStatuses())
	if err != nil {
		r.logger.Error("autoadvance: list orders failed", err)
		return
	}
	if len(orders) == 0 {
		return
	}

	for _, o := range orders {
		next, ok := transitions[o.Status]
		if !ok {
			continue
		}
		if err := r.repo.UpdateOrderStatus(ctx, o.PublicID, next); err != nil {
			r.logger.Error("autoadvance: update status failed", err,
				logger.String("order_id", o.PublicID),
				logger.String("from", o.Status),
				logger.String("to", next),
			)
			continue
		}
		_ = r.publisher.PublishJSON(ctx, events.QueueGatewayEvents, events.GatewayEvent{
			OrderID: o.PublicID,
			Status:  next,
		})
		r.logger.Info("autoadvance: order advanced",
			logger.String("order_id", o.PublicID),
			logger.String("from", o.Status),
			logger.String("to", next),
		)
	}
}
