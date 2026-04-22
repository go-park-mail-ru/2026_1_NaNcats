package restaurant

/*
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

func TestDishHandler_GetDishesByRestaurantBrandID(t *testing.T) {
	type mockInit func(uc *ucMocks.MockDishUseCase)

	tests := []struct {
		name           string
		pathID         string
		queryParams    string
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:        "Успешное получение списка блюд с параметрами",
			pathID:      "1",
			queryParams: "?limit=5&offset=10",
			mockInit: func(uc *ucMocks.MockDishUseCase) {
				dishes := []domain.Dish{
					{ID: 10, Name: "Борщ", Price: 500, ImageURL: "borsch.png"},
				}
				uc.EXPECT().
					GetDishesByRestaurantBrandID(gomock.Any(), 1, 5, 10).
					Return(dishes, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Успешное получение со значениями по умолчанию",
			pathID:      "1",
			queryParams: "",
			mockInit: func(uc *ucMocks.MockDishUseCase) {
				uc.EXPECT().
					GetDishesByRestaurantBrandID(gomock.Any(), 1, 20, 0).
					Return([]domain.Dish{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Некорректный ID бренда в пути",
			pathID:         "abc",
			queryParams:    "",
			mockInit:       func(uc *ucMocks.MockDishUseCase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "ID бренда меньше или равен нулю",
			pathID:         "0",
			queryParams:    "",
			mockInit:       func(uc *ucMocks.MockDishUseCase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Ошибка во внутреннем слое (UseCase)",
			pathID:      "1",
			queryParams: "",
			mockInit: func(uc *ucMocks.MockDishUseCase) {
				uc.EXPECT().
					GetDishesByRestaurantBrandID(gomock.Any(), 1, 20, 0).
					Return(nil, errors.New("ошибка базы данных"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			uc := ucMocks.NewMockDishUseCase(ctrl)
			l := domainMocks.NewNopLogger()
			h := NewDishHandler(uc, l)

			url := "/restaurants/brands/" + tt.pathID + "/dishes" + tt.queryParams
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.SetPathValue("id", tt.pathID)

			w := httptest.NewRecorder()

			tt.mockInit(uc)

			h.GetDishesByRestaurantBrandID(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
*/
