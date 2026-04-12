package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	domainMocks "github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain/mocks"
	ucMocks "github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRestaurantBrandHandler_GetRestaurantBrandsList(t *testing.T) {
	type mockInit func(uc *ucMocks.MockRestaurantBrandUseCase)

	tests := []struct {
		name           string
		queryParams    string
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:        "Успешное получение списка с параметрами по умолчанию",
			queryParams: "",
			mockInit: func(uc *ucMocks.MockRestaurantBrandUseCase) {
				brands := []domain.RestaurantBrand{
					{ID: 1, Name: "KFC", Description: "Wings"},
				}
				uc.EXPECT().
					GetRestaurantBrandsList(gomock.Any(), 20, 0).
					Return(brands, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Успешное получение списка с кастомными лимитом и смещением",
			queryParams: "?limit=5&offset=10",
			mockInit: func(uc *ucMocks.MockRestaurantBrandUseCase) {
				uc.EXPECT().
					GetRestaurantBrandsList(gomock.Any(), 5, 10).
					Return([]domain.RestaurantBrand{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Ошибка на стороне UseCase",
			queryParams: "",
			mockInit: func(uc *ucMocks.MockRestaurantBrandUseCase) {
				uc.EXPECT().
					GetRestaurantBrandsList(gomock.Any(), 20, 0).
					Return(nil, errors.New("ошибка базы"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			uc := ucMocks.NewMockRestaurantBrandUseCase(ctrl)
			l := domainMocks.NewNopLogger()
			h := NewRestaurantBrandHandler(uc, l)

			req := httptest.NewRequest(http.MethodGet, "/restaurants/brands"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			tt.mockInit(uc)

			h.GetRestaurantBrandsList(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRestaurantBrandHandler_GetRestaurantBrandByID(t *testing.T) {
	type mockInit func(uc *ucMocks.MockRestaurantBrandUseCase)

	tests := []struct {
		name           string
		pathID         string
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:   "Успешное получение бренда по ID",
			pathID: "1",
			mockInit: func(uc *ucMocks.MockRestaurantBrandUseCase) {
				brand := domain.RestaurantBrand{ID: 1, Name: "KFC", LogoURL: "kfc.png"}
				uc.EXPECT().
					GetRestaurantBrandByID(gomock.Any(), 1).
					Return(brand, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Ресторан не найден",
			pathID: "404",
			mockInit: func(uc *ucMocks.MockRestaurantBrandUseCase) {
				uc.EXPECT().
					GetRestaurantBrandByID(gomock.Any(), 404).
					Return(domain.RestaurantBrand{}, errors.New("not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			uc := ucMocks.NewMockRestaurantBrandUseCase(ctrl)
			l := domainMocks.NewNopLogger()
			h := NewRestaurantBrandHandler(uc, l)

			req := httptest.NewRequest(http.MethodGet, "/restaurants/brands/"+tt.pathID, nil)
			req.SetPathValue("id", tt.pathID)
			w := httptest.NewRecorder()

			tt.mockInit(uc)

			h.GetRestaurantBrandByID(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
