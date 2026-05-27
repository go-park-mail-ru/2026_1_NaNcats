package analytics

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/analyticsclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/analyticsclient/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAnalyticsHandler_GetOwnerAnalytics(t *testing.T) {
	type setup func(m *mocks.MockAnalyticsClient)

	tests := []struct {
		name           string
		url            string
		setup          setup
		expectedStatus int
	}{
		{
			name:           "Нет restaurant_id — 400",
			url:            "/api/owner/analytics?start_time=2026-01-01&end_time=2026-02-01",
			setup:          func(m *mocks.MockAnalyticsClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "restaurant_id=0 — 400",
			url:            "/api/owner/analytics?restaurant_id=0&start_time=2026-01-01&end_time=2026-02-01",
			setup:          func(m *mocks.MockAnalyticsClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Нет start_time — 400",
			url:            "/api/owner/analytics?restaurant_id=1&end_time=2026-02-01",
			setup:          func(m *mocks.MockAnalyticsClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Невалидный формат start_time — 400",
			url:            "/api/owner/analytics?restaurant_id=1&start_time=01.01.2026&end_time=2026-02-01",
			setup:          func(m *mocks.MockAnalyticsClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Невалидный формат end_time — 400",
			url:            "/api/owner/analytics?restaurant_id=1&start_time=2026-01-01&end_time=tomorrow",
			setup:          func(m *mocks.MockAnalyticsClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Ошибка gRPC — 500",
			url:  "/api/owner/analytics?restaurant_id=1&start_time=2026-01-01&end_time=2026-02-01",
			setup: func(m *mocks.MockAnalyticsClient) {
				m.EXPECT().GetOwnerStats(gomock.Any(), int64(1), gomock.Any(), gomock.Any()).
					Return(analyticsclient.OwnerStats{}, errors.New("boom"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Успех",
			url:  "/api/owner/analytics?restaurant_id=1&start_time=2026-01-01&end_time=2026-02-01",
			setup: func(m *mocks.MockAnalyticsClient) {
				m.EXPECT().GetOwnerStats(gomock.Any(), int64(1), gomock.Any(), gomock.Any()).
					Return(analyticsclient.OwnerStats{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ac := mocks.NewMockAnalyticsClient(ctrl)
			tt.setup(ac)

			h := NewAnalyticsHandler(ac, logger.NewNopLogger())
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			h.GetOwnerAnalytics(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestNewAnalyticsHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ac := mocks.NewMockAnalyticsClient(ctrl)
	h := NewAnalyticsHandler(ac, logger.NewNopLogger())
	assert.NotNil(t, h)
	assert.Equal(t, ac, h.analyticsClient)
}
