package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain/mocks"
	ucMocks "github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCSRFMiddleware_Check(t *testing.T) {
	type mockInit func(u *ucMocks.MockSessionUseCase)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		method         string
		hasCookie      bool
		headerToken    string
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:           "Успех: GET метод без проверки",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Ошибка: POST без куки",
			method:         http.MethodPost,
			hasCookie:      false,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:        "Успех: Валидный токен",
			method:      http.MethodPost,
			hasCookie:   true,
			headerToken: "valid-token",
			mockInit: func(m *ucMocks.MockSessionUseCase) {
				m.EXPECT().GetCSRF(gomock.Any(), gomock.Any()).Return("valid-token", nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSessionUC := ucMocks.NewMockSessionUseCase(ctrl)
			if tt.mockInit != nil {
				tt.mockInit(mockSessionUC)
			}

			mw := NewCSRFMiddleware(mockSessionUC, mocks.NewNopLogger())
			req := httptest.NewRequest(tt.method, "/api/test", nil)
			if tt.hasCookie {
				req.AddCookie(&http.Cookie{Name: "session_id", Value: uuid.NewString()})
			}
			if tt.headerToken != "" {
				req.Header.Set("X-CSRF-Token", tt.headerToken)
			}

			rec := httptest.NewRecorder()
			mw.Check(nextHandler).ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}
