package postgres

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/jackc/pgx/v5"
)

type tracerKey string

const (
	startTimeKey tracerKey = "db_start"
	sqlQueryKey  tracerKey = "db_sql"
)

// Трейсер походов в базу данных
type DBTracer struct {
	logger domain.Logger
}

func NewDBTracer(logger domain.Logger) *DBTracer {
	return &DBTracer{
		logger: logger,
	}
}

// pgx выполняет перед началом бд запроса
func (t *DBTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	ctx = context.WithValue(ctx, sqlQueryKey, data.SQL)
	ctx = context.WithValue(ctx, startTimeKey, time.Now())
	return ctx
}

// pgx выполняет после бд запроса
func (t *DBTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	start, ok := ctx.Value(startTimeKey).(time.Time)
	duration := time.Duration(0)
	if ok {
		duration = time.Since(start)
	}

	sqlQuery, ok := ctx.Value(sqlQueryKey).(string)

	l := t.logger.WithContext(ctx)

	fields := map[string]any{
		"sql":      sqlQuery,
		"duration": duration.String(),
	}

	if data.Err != nil {
		l.Error("sql query failed", data.Err, fields)
	} else {
		l.Debug("sql query successful", fields)
	}
}
