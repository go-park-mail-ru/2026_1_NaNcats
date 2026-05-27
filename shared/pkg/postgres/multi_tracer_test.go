package postgres

import (
	"context"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/postgres/mocks"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type contextKey string

const (
	tracer1Key contextKey = "tracer_1"
	tracer2Key contextKey = "tracer_2"
)

func TestMultiTracer_TraceQueryStart(t *testing.T) {
	type mockInit func(t1 *mocks.MockQueryTracer, t2 *mocks.MockQueryTracer)

	tests := []struct {
		name     string
		data     pgx.TraceQueryStartData
		mockInit mockInit
	}{
		{
			name: "Успешный вызов цепочки трейсеров",
			data: pgx.TraceQueryStartData{SQL: "SELECT * FROM test"},
			mockInit: func(t1 *mocks.MockQueryTracer, t2 *mocks.MockQueryTracer) {
				// Первый трейсер получает исходный контекст
				t1.EXPECT().TraceQueryStart(gomock.Any(), nil, pgx.TraceQueryStartData{SQL: "SELECT * FROM test"}).
					DoAndReturn(func(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
						return context.WithValue(ctx, tracer1Key, true)
					})

				// Второй трейсер получает контекст от первого
				t2.EXPECT().TraceQueryStart(gomock.Any(), nil, pgx.TraceQueryStartData{SQL: "SELECT * FROM test"}).
					DoAndReturn(func(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
						assert.Equal(t, true, ctx.Value(tracer1Key))
						return context.WithValue(ctx, tracer2Key, true)
					})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tracer1 := mocks.NewMockQueryTracer(ctrl)
			tracer2 := mocks.NewMockQueryTracer(ctrl)
			tt.mockInit(tracer1, tracer2)

			multiTracer := NewMultiTracer(tracer1, tracer2)

			ctx := context.Background()
			newCtx := multiTracer.TraceQueryStart(ctx, nil, tt.data)

			assert.NotNil(t, newCtx)
			assert.Equal(t, true, newCtx.Value(tracer2Key))
		})
	}
}

func TestMultiTracer_TraceQueryEnd(t *testing.T) {
	type mockInit func(t1 *mocks.MockQueryTracer, t2 *mocks.MockQueryTracer)

	tests := []struct {
		name     string
		data     pgx.TraceQueryEndData
		mockInit mockInit
	}{
		{
			name: "Успешное завершение вызова цепочки трейсеров",
			data: pgx.TraceQueryEndData{Err: nil},
			mockInit: func(t1 *mocks.MockQueryTracer, t2 *mocks.MockQueryTracer) {
				t1.EXPECT().TraceQueryEnd(gomock.Any(), nil, pgx.TraceQueryEndData{Err: nil})
				t2.EXPECT().TraceQueryEnd(gomock.Any(), nil, pgx.TraceQueryEndData{Err: nil})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tracer1 := mocks.NewMockQueryTracer(ctrl)
			tracer2 := mocks.NewMockQueryTracer(ctrl)
			tt.mockInit(tracer1, tracer2)

			multiTracer := NewMultiTracer(tracer1, tracer2)

			multiTracer.TraceQueryEnd(context.Background(), nil, tt.data)
		})
	}
}
