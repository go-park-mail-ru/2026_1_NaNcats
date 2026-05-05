package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	loggerMocks "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger/mocks"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestDBTracer_TraceQueryStart(t *testing.T) {
	tests := []struct {
		name string
		data pgx.TraceQueryStartData
	}{
		{
			name: "Успешное обогащение контекста",
			data: pgx.TraceQueryStartData{SQL: "SELECT * FROM users"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := loggerMocks.NewMockLogger(ctrl)
			tracer := NewDBTracer(mockLogger)

			ctx := context.Background()
			newCtx := tracer.TraceQueryStart(ctx, nil, tt.data)

			// Assert
			require.NotNil(t, newCtx)

			sqlVal := newCtx.Value(sqlQueryKey)
			assert.Equal(t, tt.data.SQL, sqlVal)

			startVal := newCtx.Value(startTimeKey)
			assert.IsType(t, time.Time{}, startVal)
		})
	}
}

func TestDBTracer_TraceQueryEnd(t *testing.T) {
	type mockInit func(m *loggerMocks.MockLogger)

	tests := []struct {
		name     string
		ctxSetup func() context.Context
		data     pgx.TraceQueryEndData
		mockInit mockInit
	}{
		{
			name: "Успешный запрос, вызывается Debug",
			ctxSetup: func() context.Context {
				ctx := context.WithValue(context.Background(), sqlQueryKey, "SELECT 1")
				return context.WithValue(ctx, startTimeKey, time.Now().Add(-10*time.Millisecond))
			},
			data: pgx.TraceQueryEndData{Err: nil},
			mockInit: func(m *loggerMocks.MockLogger) {
				m.EXPECT().WithContext(gomock.Any()).Return(m)
				m.EXPECT().Debug("sql query successful", gomock.Any(), gomock.Any())
			},
		},
		{
			name: "Запрос с ошибкой, вызывается Error",
			ctxSetup: func() context.Context {
				ctx := context.WithValue(context.Background(), sqlQueryKey, "INSERT INTO err")
				return context.WithValue(ctx, startTimeKey, time.Now())
			},
			data: pgx.TraceQueryEndData{Err: errors.New("db error")},
			mockInit: func(m *loggerMocks.MockLogger) {
				m.EXPECT().WithContext(gomock.Any()).Return(m)
				m.EXPECT().Error("sql query failed", errors.New("db error"), gomock.Any(), gomock.Any())
			},
		},
		{
			name: "Контекст без значений (дефолтные параметры)",
			ctxSetup: func() context.Context {
				return context.Background()
			},
			data: pgx.TraceQueryEndData{Err: nil},
			mockInit: func(m *loggerMocks.MockLogger) {
				m.EXPECT().WithContext(gomock.Any()).Return(m)
				// Проверяем, что sqlQuery откатывается к "unknown sql"
				m.EXPECT().Debug("sql query successful", logger.String("sql", "unknown sql"), gomock.Any())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := loggerMocks.NewMockLogger(ctrl)
			tt.mockInit(mockLogger)

			tracer := NewDBTracer(mockLogger)
			ctx := tt.ctxSetup()

			tracer.TraceQueryEnd(ctx, nil, tt.data)
		})
	}
}
