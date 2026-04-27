package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type MultiTracer struct {
	tracers []pgx.QueryTracer
}

func NewMultiTracer(tracers ...pgx.QueryTracer) *MultiTracer {
	return &MultiTracer{tracers: tracers}
}

func (m *MultiTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	for _, t := range m.tracers {
		ctx = t.TraceQueryStart(ctx, conn, data)
	}
	return ctx
}

func (m *MultiTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	for _, t := range m.tracers {
		t.TraceQueryEnd(ctx, conn, data)
	}
}
