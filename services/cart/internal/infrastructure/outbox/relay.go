package outbox

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq"
	gwEvent "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type outboxEventDB struct {
	ID          string `db:"id"`
	AggregateID string `db:"aggregate_id"`
	EventType   string `db:"event_type"`
	Payload     []byte `db:"payload"`
}

type Relay struct {
	db          *pgxpool.Pool
	rabbit      *rabbitmq.RabbitClient
	logger      logger.Logger
	targetQueue string
}

func NewRelay(db *pgxpool.Pool, rabbit *rabbitmq.RabbitClient, l logger.Logger, q string) *Relay {
	return &Relay{
		db:          db,
		rabbit:      rabbit,
		logger:      l,
		targetQueue: q,
	}
}

func (r *Relay) Run(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	r.logger.Info("Started Outbox Relay for Cart Service")

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("Outbox Relay stopped")
			return
		case <-ticker.C:
			r.processEvents(ctx)
		}
	}
}

func (r *Relay) processEvents(ctx context.Context) {
	query := `
		SELECT id, aggregate_id, event_type, payload 
		FROM outbox_events 
		WHERE status = 'PENDING' 
		ORDER BY created_at ASC 
		LIMIT 50
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		r.logger.Error("Failed to query outbox events", err)
		return
	}

	events, err := pgx.CollectRows(rows, pgx.RowToStructByName[outboxEventDB])
	if err != nil {
		r.logger.Error("Failed to collect outbox rows", err)
		return
	}

	if len(events) == 0 {
		return // Нет новых событий
	}

	var processedIDs []string

	for _, event := range events {
		gatewayMsg := gwEvent.GatewayEvent{
			CartID:    event.AggregateID,
			EventType: event.EventType,
			Payload:   event.Payload,
		}

		err = r.rabbit.PublishJSON(ctx, r.targetQueue, gatewayMsg)
		if err != nil {
			r.logger.Error("Failed to publish event to RabbitMQ", err, logger.String("event_id", event.ID))
			break
		}

		processedIDs = append(processedIDs, event.ID)
	}

	if len(processedIDs) > 0 {
		_, err = r.db.Exec(ctx, `UPDATE outbox_events SET status = 'PROCESSED' WHERE id = ANY($1)`, processedIDs)
		if err != nil {
			r.logger.Error("Failed to bulk update outbox events", err)
		}
	}
}
