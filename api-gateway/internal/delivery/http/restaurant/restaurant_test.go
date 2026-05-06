package restaurant

import (
	json "encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/restaurantclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/restaurantclient/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/restaurant"
	pbRestaurant "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/restaurant"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRestaurantHandler_GetRestaurantBrandsList(t *testing.T) {
	type mockInit func(m *mocks.MockRestaurantClient)

	tests := []struct {
		name           string
		queryParams    string
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:        "Успешное получение списка с параметрами по умолчанию",
			queryParams: "",
			mockInit: func(m *mocks.MockRestaurantClient) {
				m.EXPECT().
					GetRestaurantBrandsList(gomock.Any(), int32(20), int32(0)).
					Return([]*pb.RestaurantBrand{
						{Id: 1, Name: "Бренд 1", LogoUrl: "logo.png"},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Успешное получение списка с кастомными лимитом и смещением",
			queryParams: "?limit=10&offset=5",
			mockInit: func(m *mocks.MockRestaurantClient) {
				m.EXPECT().
					GetRestaurantBrandsList(gomock.Any(), int32(10), int32(5)).
					Return([]*pb.RestaurantBrand{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Ошибка gRPC клиента",
			queryParams: "",
			mockInit: func(m *mocks.MockRestaurantClient) {
				m.EXPECT().
					GetRestaurantBrandsList(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("grpc error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockRestaurantClient(ctrl)
			tt.mockInit(mockClient)

			h := NewRestaurantHandler(mockClient, logger.NewNopLogger())

			req := httptest.NewRequest(http.MethodGet, "/api/restaurants/brands"+tt.queryParams, nil)
			rec := httptest.NewRecorder()

			h.GetRestaurantBrandsList(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestRestaurantHandler_GetRestaurantBrandByID(t *testing.T) {
	type mockInit func(m *mocks.MockRestaurantClient)

	tests := []struct {
		name           string
		idPath         string
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:   "Успешное получение бренда по ID",
			idPath: "1",
			mockInit: func(m *mocks.MockRestaurantClient) {
				m.EXPECT().
					GetRestaurantBrandByID(gomock.Any(), int64(1)).
					Return(&pb.RestaurantBrand{Id: 1, Name: "Тест"}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Ошибка: невалидный формат ID",
			idPath:         "abc",
			mockInit:       func(m *mocks.MockRestaurantClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Ошибка: ID меньше или равен нулю",
			idPath:         "0",
			mockInit:       func(m *mocks.MockRestaurantClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Ошибка: ресторан не найден",
			idPath: "404",
			mockInit: func(m *mocks.MockRestaurantClient) {
				m.EXPECT().
					GetRestaurantBrandByID(gomock.Any(), int64(404)).
					Return(nil, restaurantclient.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "Внутренняя ошибка сервиса",
			idPath: "500",
			mockInit: func(m *mocks.MockRestaurantClient) {
				m.EXPECT().
					GetRestaurantBrandByID(gomock.Any(), int64(500)).
					Return(nil, errors.New("fatal"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockRestaurantClient(ctrl)
			tt.mockInit(mockClient)

			h := NewRestaurantHandler(mockClient, logger.NewNopLogger())

			req := httptest.NewRequest(http.MethodGet, "/api/restaurants/brands/"+tt.idPath, nil)
			req.SetPathValue("id", tt.idPath)
			rec := httptest.NewRecorder()

			h.GetRestaurantBrandByID(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestRestaurantHandler_GetDishesByRestaurantBrandID(t *testing.T) {
	type mockInit func(m *mocks.MockRestaurantClient)

	tests := []struct {
		name           string
		idPath         string
		queryParams    string
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:        "Успешное получение списка блюд без поиска",
			idPath:      "1",
			queryParams: "?limit=10&offset=5",
			mockInit: func(m *mocks.MockRestaurantClient) {
				m.EXPECT().
					GetDishesByRestaurantBrandID(gomock.Any(), int64(1), int32(10), int32(5)).
					Return([]*pbRestaurant.Dish{
						{Id: 101, Name: "Борщ", Price: 300},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Успешный поиск блюд внутри ресторана",
			idPath:      "1",
			queryParams: "?q=pizza&limit=5",
			mockInit: func(m *mocks.MockRestaurantClient) {
				m.EXPECT().
					SearchDishesByBrand(gomock.Any(), int64(1), "pizza", int32(5)).
					Return([]*pbRestaurant.Dish{
						{Id: 202, Name: "Margarita", Price: 500},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Ошибка: невалидный ID бренда",
			idPath:         "invalid",
			mockInit:       func(m *mocks.MockRestaurantClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Ошибка: ID бренда меньше единицы",
			idPath:         "0",
			mockInit:       func(m *mocks.MockRestaurantClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Ошибка: блюда или ресторан не найдены",
			idPath: "404",
			mockInit: func(m *mocks.MockRestaurantClient) {
				m.EXPECT().
					GetDishesByRestaurantBrandID(gomock.Any(), int64(404), gomock.Any(), gomock.Any()).
					Return(nil, restaurantclient.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "Внутренняя ошибка сервиса",
			idPath: "1",
			mockInit: func(m *mocks.MockRestaurantClient) {
				m.EXPECT().
					GetDishesByRestaurantBrandID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("grpc crash"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockRestaurantClient(ctrl)
			tt.mockInit(mockClient)

			h := NewRestaurantHandler(mockClient, logger.NewNopLogger())

			req := httptest.NewRequest(http.MethodGet, "/api/restaurants/brands/"+tt.idPath+"/dishes"+tt.queryParams, nil)
			req.SetPathValue("id", tt.idPath)
			rec := httptest.NewRecorder()

			h.GetDishesByRestaurantBrandID(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestRestaurantHandler_GetCategories(t *testing.T) {
	h := NewRestaurantHandler(nil, logger.NewNopLogger())

	t.Run("Успешное получение списка категорий", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/restaurants/categories", nil)
		rec := httptest.NewRecorder()

		h.GetCategories(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")

		var resp CategoriesResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.Categories)
	})
}

func TestRestaurantHandler_GetRestaurantBrandsListByCategory(t *testing.T) {
	type mockInit func(m *mocks.MockRestaurantClient)

	tests := []struct {
		name           string
		slug           string
		queryParams    string
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:        "Успешное получение по известной категории",
			slug:        "pizza", // Из categoryNameMap
			queryParams: "?limit=10",
			mockInit: func(m *mocks.MockRestaurantClient) {
				m.EXPECT().
					GetRestaurantBrandsListByCategoryName(gomock.Any(), "Пицца", int32(10), int32(0)).
					Return([]*pb.RestaurantBrand{{Id: 1, Name: "Pizza Dominos"}}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Успешный фолбэк при неизвестной категории",
			slug:        "unknown-slug",
			queryParams: "?offset=5",
			mockInit: func(m *mocks.MockRestaurantClient) {
				m.EXPECT().
					GetRestaurantBrandsList(gomock.Any(), int32(20), int32(5)).
					Return([]*pb.RestaurantBrand{{Id: 2, Name: "Any Rest"}}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Ошибка gRPC при получении по имени категории",
			slug:        "burgers",
			queryParams: "",
			mockInit: func(m *mocks.MockRestaurantClient) {
				m.EXPECT().
					GetRestaurantBrandsListByCategoryName(gomock.Any(), "Бургеры", int32(20), int32(0)).
					Return(nil, errors.New("grpc error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Ошибка gRPC при фолбэке",
			slug: "ghost",
			mockInit: func(m *mocks.MockRestaurantClient) {
				m.EXPECT().
					GetRestaurantBrandsList(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("fatal"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockRestaurantClient(ctrl)
			tt.mockInit(mockClient)

			h := NewRestaurantHandler(mockClient, logger.NewNopLogger())

			req := httptest.NewRequest(http.MethodGet, "/api/restaurants/categories/"+tt.slug+"/brands"+tt.queryParams, nil)
			req.SetPathValue("slug", tt.slug)
			rec := httptest.NewRecorder()

			h.GetRestaurantBrandsListByCategory(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestRestaurantHandler_SearchRestaurants(t *testing.T) {
	type mockInit func(m *mocks.MockRestaurantClient)

	tests := []struct {
		name           string
		queryParams    string
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:        "Успешный глобальный поиск",
			queryParams: "?q=бургер",
			mockInit: func(m *mocks.MockRestaurantClient) {
				m.EXPECT().
					SearchRestaurantBrands(gomock.Any(), "бургер", int32(20), int32(0)).
					Return([]*pb.RestaurantBrand{{Id: 1, Name: "Burger Club"}}, nil)
				m.EXPECT().
					SearchDishes(gomock.Any(), "бургер", int32(20)).
					Return([]*pb.Dish{{Id: 101, Name: "Чизбургер", RestaurantBrandId: 1}}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Поиск только с ресторанами (блюда упали, но статус 200)",
			queryParams: "?q=test",
			mockInit: func(m *mocks.MockRestaurantClient) {
				m.EXPECT().
					SearchRestaurantBrands(gomock.Any(), "test", int32(20), int32(0)).
					Return([]*pb.RestaurantBrand{{Id: 1}}, nil)
				m.EXPECT().
					SearchDishes(gomock.Any(), "test", int32(20)).
					Return(nil, errors.New("dish search failed"))
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Ошибка: отсутствует поисковый запрос",
			queryParams:    "",
			mockInit:       func(m *mocks.MockRestaurantClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Успешный поиск с кастомной пагинацией",
			queryParams: "?q=pizza&limit=5&offset=10",
			mockInit: func(m *mocks.MockRestaurantClient) {
				m.EXPECT().
					SearchRestaurantBrands(gomock.Any(), "pizza", int32(5), int32(10)).
					Return([]*pb.RestaurantBrand{}, nil)
				m.EXPECT().
					SearchDishes(gomock.Any(), "pizza", int32(5)).
					Return([]*pb.Dish{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockRestaurantClient(ctrl)
			tt.mockInit(mockClient)

			h := NewRestaurantHandler(mockClient, logger.NewNopLogger())

			req := httptest.NewRequest(http.MethodGet, "/api/restaurants/search"+tt.queryParams, nil)
			rec := httptest.NewRecorder()

			h.SearchRestaurants(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}
