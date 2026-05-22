package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/domain"
	ucMocks "github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/usecase/mocks"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/restaurant"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestRestaurantHandler_GetRestaurantBrandsList(t *testing.T) {
	type mockInit func(buc *ucMocks.MockRestaurantBrandUseCase)

	tests := []struct {
		name           string
		req            *pb.GetRestaurantBrandsListRequest
		mockInit       mockInit
		expectedStatus codes.Code
	}{
		{
			name: "Успешное получение списка по умолчанию",
			req:  &pb.GetRestaurantBrandsListRequest{Limit: 10, Offset: 0},
			mockInit: func(buc *ucMocks.MockRestaurantBrandUseCase) {
				buc.EXPECT().GetRestaurantBrandsList(gomock.Any(), 10, 0).Return([]domain.RestaurantBrand{{ID: 1, Name: "Rest"}}, nil)
			},
			expectedStatus: codes.OK,
		},
		{
			name: "Ошибка в UseCase при получении списка",
			req:  &pb.GetRestaurantBrandsListRequest{Limit: 10, Offset: 0},
			mockInit: func(buc *ucMocks.MockRestaurantBrandUseCase) {
				buc.EXPECT().GetRestaurantBrandsList(gomock.Any(), 10, 0).Return(nil, errors.New("failed"))
			},
			expectedStatus: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			buc := ucMocks.NewMockRestaurantBrandUseCase(ctrl)
			h := NewRestaurantHandler(nil, buc, nil)

			tt.mockInit(buc)

			resp, err := h.GetRestaurantBrandsList(context.Background(), tt.req)

			if tt.expectedStatus == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Len(t, resp.RestaurantBrands, 1)
			} else {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedStatus, st.Code())
			}
		})
	}
}

func TestRestaurantHandler_GetRestaurantBrandsByCategory(t *testing.T) {
	type mockInit func(buc *ucMocks.MockRestaurantBrandUseCase)

	tests := []struct {
		name           string
		req            *pb.GetRestaurantBrandsByCategoryRequest
		mockInit       mockInit
		expectedStatus codes.Code
	}{
		{
			name: "Успешная фильтрация по категории",
			req:  &pb.GetRestaurantBrandsByCategoryRequest{CategoryName: "Пицца", Limit: 5, Offset: 0},
			mockInit: func(buc *ucMocks.MockRestaurantBrandUseCase) {
				buc.EXPECT().GetRestaurantBrandsByCategoryName(gomock.Any(), "Пицца", 5, 0).Return([]domain.RestaurantBrand{{ID: 2, Name: "Pizzeria"}}, nil)
			},
			expectedStatus: codes.OK,
		},
		{
			name: "Ошибка UseCase при получении по категории",
			req:  &pb.GetRestaurantBrandsByCategoryRequest{CategoryName: "Бургеры", Limit: 10, Offset: 0},
			mockInit: func(buc *ucMocks.MockRestaurantBrandUseCase) {
				buc.EXPECT().GetRestaurantBrandsByCategoryName(gomock.Any(), "Бургеры", 10, 0).Return(nil, errors.New("internal"))
			},
			expectedStatus: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			buc := ucMocks.NewMockRestaurantBrandUseCase(ctrl)
			h := NewRestaurantHandler(nil, buc, nil)

			tt.mockInit(buc)

			resp, err := h.GetRestaurantBrandsByCategory(context.Background(), tt.req)

			if tt.expectedStatus == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			} else {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedStatus, st.Code())
			}
		})
	}
}

func TestRestaurantHandler_SearchRestaurantBrands(t *testing.T) {
	type mockInit func(buc *ucMocks.MockRestaurantBrandUseCase)

	tests := []struct {
		name           string
		req            *pb.SearchRestaurantBrandsRequest
		mockInit       mockInit
		expectedStatus codes.Code
	}{
		{
			name: "Успешный поиск ресторанов",
			req:  &pb.SearchRestaurantBrandsRequest{Query: "бургер", Limit: 10, Offset: 0},
			mockInit: func(buc *ucMocks.MockRestaurantBrandUseCase) {
				buc.EXPECT().SearchRestaurantBrands(gomock.Any(), "бургер", 10, 0).Return([]domain.RestaurantBrand{{ID: 1, Name: "Burger"}}, nil)
			},
			expectedStatus: codes.OK,
		},
		{
			name: "Ошибка во время поиска",
			req:  &pb.SearchRestaurantBrandsRequest{Query: "test", Limit: 10, Offset: 0},
			mockInit: func(buc *ucMocks.MockRestaurantBrandUseCase) {
				buc.EXPECT().SearchRestaurantBrands(gomock.Any(), "test", 10, 0).Return(nil, errors.New("db error"))
			},
			expectedStatus: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			buc := ucMocks.NewMockRestaurantBrandUseCase(ctrl)
			h := NewRestaurantHandler(nil, buc, nil)

			tt.mockInit(buc)

			resp, err := h.SearchRestaurantBrands(context.Background(), tt.req)

			if tt.expectedStatus == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			} else {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedStatus, st.Code())
			}
		})
	}
}

func TestRestaurantHandler_GetRestaurantBrandByID(t *testing.T) {
	type mockInit func(buc *ucMocks.MockRestaurantBrandUseCase)

	tests := []struct {
		name           string
		req            *pb.GetRestaurantBrandByIDRequest
		mockInit       mockInit
		expectedStatus codes.Code
	}{
		{
			name: "Успешное получение по ID",
			req:  &pb.GetRestaurantBrandByIDRequest{Id: 1},
			mockInit: func(buc *ucMocks.MockRestaurantBrandUseCase) {
				buc.EXPECT().GetRestaurantBrandByID(gomock.Any(), int64(1)).Return(domain.RestaurantBrand{ID: 1, Name: "KFC"}, nil)
			},
			expectedStatus: codes.OK,
		},
		{
			name: "Ресторан не найден",
			req:  &pb.GetRestaurantBrandByIDRequest{Id: 404},
			mockInit: func(buc *ucMocks.MockRestaurantBrandUseCase) {
				buc.EXPECT().GetRestaurantBrandByID(gomock.Any(), int64(404)).Return(domain.RestaurantBrand{}, domain.ErrRestaurantNotFound)
			},
			expectedStatus: codes.NotFound,
		},
		{
			name: "Системная ошибка сервера",
			req:  &pb.GetRestaurantBrandByIDRequest{Id: 500},
			mockInit: func(buc *ucMocks.MockRestaurantBrandUseCase) {
				buc.EXPECT().GetRestaurantBrandByID(gomock.Any(), int64(500)).Return(domain.RestaurantBrand{}, errors.New("internal"))
			},
			expectedStatus: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			buc := ucMocks.NewMockRestaurantBrandUseCase(ctrl)
			h := NewRestaurantHandler(nil, buc, nil)

			tt.mockInit(buc)

			resp, err := h.GetRestaurantBrandByID(context.Background(), tt.req)

			if tt.expectedStatus == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.req.Id, resp.RestaurantBrand.Id)
			} else {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedStatus, st.Code())
			}
		})
	}
}

func TestRestaurantHandler_GetRestaurantBrandsByIDs(t *testing.T) {
	type mockInit func(buc *ucMocks.MockRestaurantBrandUseCase)

	tests := []struct {
		name           string
		req            *pb.GetRestaurantBrandsByIDsRequest
		mockInit       mockInit
		expectedStatus codes.Code
	}{
		{
			name: "Успешное пакетное получение брендов",
			req:  &pb.GetRestaurantBrandsByIDsRequest{BrandIds: []int64{1, 2}},
			mockInit: func(buc *ucMocks.MockRestaurantBrandUseCase) {
				buc.EXPECT().GetRestaurantBrandsByIDs(gomock.Any(), []int64{1, 2}).
					Return([]domain.RestaurantBrand{{ID: 1}, {ID: 2}}, nil)
			},
			expectedStatus: codes.OK,
		},
		{
			name: "Ошибка UseCase при пакетном получении",
			req:  &pb.GetRestaurantBrandsByIDsRequest{BrandIds: []int64{1}},
			mockInit: func(buc *ucMocks.MockRestaurantBrandUseCase) {
				buc.EXPECT().GetRestaurantBrandsByIDs(gomock.Any(), []int64{1}).
					Return(nil, errors.New("db error"))
			},
			expectedStatus: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			buc := ucMocks.NewMockRestaurantBrandUseCase(ctrl)
			h := NewRestaurantHandler(nil, buc, nil)

			tt.mockInit(buc)

			resp, err := h.GetRestaurantBrandsByIDs(context.Background(), tt.req)

			if tt.expectedStatus == codes.OK {
				assert.NoError(t, err)
				assert.Len(t, resp.RestaurantBrands, len(tt.req.BrandIds))
			} else {
				st, _ := status.FromError(err)
				assert.Equal(t, tt.expectedStatus, st.Code())
			}
		})
	}
}

func TestRestaurantHandler_GetDishesByRestaurantBrandID(t *testing.T) {
	type mockInit func(duc *ucMocks.MockDishUseCase)

	tests := []struct {
		name           string
		req            *pb.GetDishesByRestaurantBrandIDRequest
		mockInit       mockInit
		expectedStatus codes.Code
	}{
		{
			name: "Успешное получение блюд",
			req:  &pb.GetDishesByRestaurantBrandIDRequest{RestaurantBrandId: 1, Limit: 10, Offset: 0},
			mockInit: func(duc *ucMocks.MockDishUseCase) {
				duc.EXPECT().GetDishesByRestaurantBrandID(gomock.Any(), int64(1), 10, 0).
					Return([]domain.Dish{{ID: 101, Name: "Pasta"}}, nil)
			},
			expectedStatus: codes.OK,
		},
		{
			name: "Ошибка UseCase при получении блюд",
			req:  &pb.GetDishesByRestaurantBrandIDRequest{RestaurantBrandId: 1},
			mockInit: func(duc *ucMocks.MockDishUseCase) {
				duc.EXPECT().GetDishesByRestaurantBrandID(gomock.Any(), int64(1), 0, 0).
					Return(nil, domain.ErrDishNotFound)
			},
			expectedStatus: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			duc := ucMocks.NewMockDishUseCase(ctrl)
			h := NewRestaurantHandler(nil, nil, duc)

			tt.mockInit(duc)

			resp, err := h.GetDishesByRestaurantBrandID(context.Background(), tt.req)

			if tt.expectedStatus == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			} else {
				st, _ := status.FromError(err)
				assert.Equal(t, tt.expectedStatus, st.Code())
			}
		})
	}
}

func TestRestaurantHandler_SearchDishesByBrand(t *testing.T) {
	type mockInit func(duc *ucMocks.MockDishUseCase)

	tests := []struct {
		name           string
		req            *pb.SearchDishesByBrandRequest
		mockInit       mockInit
		expectedStatus codes.Code
	}{
		{
			name: "Успешный поиск блюд в ресторане",
			req:  &pb.SearchDishesByBrandRequest{RestaurantBrandId: 1, Query: "пицца", Limit: 5},
			mockInit: func(duc *ucMocks.MockDishUseCase) {
				duc.EXPECT().SearchDishesByBrand(gomock.Any(), int64(1), "пицца", 5).
					Return([]domain.Dish{{ID: 202, Name: "Pizza Pepperoni"}}, nil)
			},
			expectedStatus: codes.OK,
		},
		{
			name: "Ошибка UseCase при поиске блюд в ресторане",
			req:  &pb.SearchDishesByBrandRequest{RestaurantBrandId: 1, Query: "test", Limit: 10},
			mockInit: func(duc *ucMocks.MockDishUseCase) {
				duc.EXPECT().SearchDishesByBrand(gomock.Any(), int64(1), "test", 10).
					Return(nil, errors.New("search fail"))
			},
			expectedStatus: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			duc := ucMocks.NewMockDishUseCase(ctrl)
			h := NewRestaurantHandler(nil, nil, duc)

			tt.mockInit(duc)

			resp, err := h.SearchDishesByBrand(context.Background(), tt.req)

			if tt.expectedStatus == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			} else {
				st, _ := status.FromError(err)
				assert.Equal(t, tt.expectedStatus, st.Code())
			}
		})
	}
}

func TestRestaurantHandler_GetDishesByIDs(t *testing.T) {
	type mockInit func(duc *ucMocks.MockDishUseCase)

	tests := []struct {
		name           string
		req            *pb.GetDishesByIDsRequest
		mockInit       mockInit
		expectedStatus codes.Code
	}{
		{
			name: "Успешное получение блюд по списку ID",
			req:  &pb.GetDishesByIDsRequest{DishIds: []int64{1, 2}},
			mockInit: func(duc *ucMocks.MockDishUseCase) {
				duc.EXPECT().GetDishesByIDs(gomock.Any(), []int64{1, 2}).
					Return([]domain.Dish{{ID: 1, Name: "Dish 1"}, {ID: 2, Name: "Dish 2"}}, nil)
			},
			expectedStatus: codes.OK,
		},
		{
			name: "Ошибка UseCase при получении по ID",
			req:  &pb.GetDishesByIDsRequest{DishIds: []int64{1}},
			mockInit: func(duc *ucMocks.MockDishUseCase) {
				duc.EXPECT().GetDishesByIDs(gomock.Any(), []int64{1}).
					Return(nil, errors.New("grpc error"))
			},
			expectedStatus: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			duc := ucMocks.NewMockDishUseCase(ctrl)
			h := NewRestaurantHandler(nil, nil, duc)

			tt.mockInit(duc)

			resp, err := h.GetDishesByIDs(context.Background(), tt.req)

			if tt.expectedStatus == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			} else {
				st, _ := status.FromError(err)
				assert.Equal(t, tt.expectedStatus, st.Code())
			}
		})
	}
}

func TestRestaurantHandler_SearchDishes(t *testing.T) {
	type mockInit func(duc *ucMocks.MockDishUseCase)

	tests := []struct {
		name           string
		req            *pb.SearchDishesRequest
		mockInit       mockInit
		expectedStatus codes.Code
	}{
		{
			name: "Успешный глобальный поиск блюд",
			req:  &pb.SearchDishesRequest{Query: "суп", Limit: 20},
			mockInit: func(duc *ucMocks.MockDishUseCase) {
				duc.EXPECT().SearchDishes(gomock.Any(), "суп", 20).
					Return([]domain.Dish{{ID: 10, Name: "Том ям"}}, nil)
			},
			expectedStatus: codes.OK,
		},
		{
			name: "Ошибка UseCase при глобальном поиске",
			req:  &pb.SearchDishesRequest{Query: "query", Limit: 20},
			mockInit: func(duc *ucMocks.MockDishUseCase) {
				duc.EXPECT().SearchDishes(gomock.Any(), "query", 20).
					Return(nil, errors.New("db fail"))
			},
			expectedStatus: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			duc := ucMocks.NewMockDishUseCase(ctrl)
			h := NewRestaurantHandler(nil, nil, duc)

			tt.mockInit(duc)

			resp, err := h.SearchDishes(context.Background(), tt.req)

			if tt.expectedStatus == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			} else {
				st, _ := status.FromError(err)
				assert.Equal(t, tt.expectedStatus, st.Code())
			}
		})
	}
}

func TestRestaurantHandler_CreateRestaurantBrand(t *testing.T) {
	type mockInit func(buc *ucMocks.MockRestaurantBrandUseCase)

	tests := []struct {
		name           string
		req            *pb.CreateBrandRequest
		mockInit       mockInit
		expectedStatus codes.Code
	}{
		{
			name: "Успешное создание бренда",
			req: &pb.CreateBrandRequest{
				OwnerId:        1,
				Name:           "New Rest",
				Description:    "Desc",
				IdempotencyKey: "key-1",
			},
			mockInit: func(buc *ucMocks.MockRestaurantBrandUseCase) {
				expectedDomain := domain.RestaurantBrand{OwnerProfileID: 1, Name: "New Rest", Description: "Desc"}
				buc.EXPECT().CreateRestaurantBrand(gomock.Any(), expectedDomain, gomock.Any(), "key-1").
					Return(domain.RestaurantBrand{ID: 10, Name: "New Rest"}, nil)
			},
			expectedStatus: codes.OK,
		},
		{
			name: "Ошибка создания бренда (UseCase)",
			req:  &pb.CreateBrandRequest{Name: "Fail"},
			mockInit: func(buc *ucMocks.MockRestaurantBrandUseCase) {
				buc.EXPECT().CreateRestaurantBrand(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(domain.RestaurantBrand{}, errors.New("internal"))
			},
			expectedStatus: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			buc := ucMocks.NewMockRestaurantBrandUseCase(ctrl)
			h := NewRestaurantHandler(nil, buc, nil)

			tt.mockInit(buc)

			resp, err := h.CreateRestaurantBrand(context.Background(), tt.req)

			if tt.expectedStatus == codes.OK {
				assert.NoError(t, err)
				assert.Equal(t, tt.req.Name, resp.Name)
			} else {
				st, _ := status.FromError(err)
				assert.Equal(t, tt.expectedStatus, st.Code())
			}
		})
	}
}

func TestRestaurantHandler_UpdateRestaurantBrand(t *testing.T) {
	type mockInit func(buc *ucMocks.MockRestaurantBrandUseCase)

	name := "Updated Name"
	tier := int32(2)

	tests := []struct {
		name           string
		req            *pb.UpdateBrandRequest
		mockInit       mockInit
		expectedStatus codes.Code
	}{
		{
			name: "Успешное обновление всех полей",
			req: &pb.UpdateBrandRequest{
				Id:             1,
				Name:           &name,
				PromotionTier:  &tier,
				IdempotencyKey: "idem-1",
			},
			mockInit: func(buc *ucMocks.MockRestaurantBrandUseCase) {
				buc.EXPECT().UpdateRestaurantBrand(gomock.Any(), domain.RestaurantBrand{
					ID: 1, Name: name, PromotionTier: 2,
				}, gomock.Any(), "idem-1").
					Return(domain.RestaurantBrand{ID: 1, Name: name, PromotionTier: 2}, nil)
			},
			expectedStatus: codes.OK,
		},
		{
			name: "Успешное частичное обновление (только ID)",
			req:  &pb.UpdateBrandRequest{Id: 5},
			mockInit: func(buc *ucMocks.MockRestaurantBrandUseCase) {
				buc.EXPECT().UpdateRestaurantBrand(gomock.Any(), domain.RestaurantBrand{ID: 5}, gomock.Any(), gomock.Any()).
					Return(domain.RestaurantBrand{ID: 5}, nil)
			},
			expectedStatus: codes.OK,
		},
		{
			name: "Ошибка обновления (ресторан не найден)",
			req:  &pb.UpdateBrandRequest{Id: 404},
			mockInit: func(buc *ucMocks.MockRestaurantBrandUseCase) {
				buc.EXPECT().UpdateRestaurantBrand(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(domain.RestaurantBrand{}, domain.ErrRestaurantNotFound)
			},
			expectedStatus: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			buc := ucMocks.NewMockRestaurantBrandUseCase(ctrl)
			h := NewRestaurantHandler(nil, buc, nil)

			tt.mockInit(buc)

			resp, err := h.UpdateRestaurantBrand(context.Background(), tt.req)

			if tt.expectedStatus == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			} else {
				st, _ := status.FromError(err)
				assert.Equal(t, tt.expectedStatus, st.Code())
			}
		})
	}
}

func TestRestaurantHandler_DeleteRestaurantBrand(t *testing.T) {
	type mockInit func(buc *ucMocks.MockRestaurantBrandUseCase)

	tests := []struct {
		name           string
		req            *pb.DeleteBrandRequest
		mockInit       mockInit
		expectedStatus codes.Code
	}{
		{
			name: "Успешное удаление бренда",
			req:  &pb.DeleteBrandRequest{Id: 1},
			mockInit: func(buc *ucMocks.MockRestaurantBrandUseCase) {
				buc.EXPECT().DeleteRestaurantBrand(gomock.Any(), int64(1)).Return(nil)
			},
			expectedStatus: codes.OK,
		},
		{
			name: "Ошибка: бренд не найден",
			req:  &pb.DeleteBrandRequest{Id: 404},
			mockInit: func(buc *ucMocks.MockRestaurantBrandUseCase) {
				buc.EXPECT().DeleteRestaurantBrand(gomock.Any(), int64(404)).Return(domain.ErrRestaurantNotFound)
			},
			expectedStatus: codes.NotFound,
		},
		{
			name: "Внутренняя ошибка при удалении",
			req:  &pb.DeleteBrandRequest{Id: 500},
			mockInit: func(buc *ucMocks.MockRestaurantBrandUseCase) {
				buc.EXPECT().DeleteRestaurantBrand(gomock.Any(), int64(500)).Return(errors.New("db crash"))
			},
			expectedStatus: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			buc := ucMocks.NewMockRestaurantBrandUseCase(ctrl)
			h := NewRestaurantHandler(nil, buc, nil)

			tt.mockInit(buc)

			_, err := h.DeleteRestaurantBrand(context.Background(), tt.req)

			if tt.expectedStatus == codes.OK {
				assert.NoError(t, err)
			} else {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedStatus, st.Code())
			}
		})
	}
}

func TestRestaurantHandler_CreateDish(t *testing.T) {
	type mockInit func(duc *ucMocks.MockDishUseCase)

	tests := []struct {
		name           string
		req            *pb.CreateDishRequest
		mockInit       mockInit
		expectedStatus codes.Code
	}{
		{
			name: "Успешное создание блюда",
			req: &pb.CreateDishRequest{
				RestaurantBrandId: 1,
				Name:              "Борщ",
				Price:             500000000,
				Description:       "Вкусный",
				IdempotencyKey:    "dish-key-1",
			},
			mockInit: func(duc *ucMocks.MockDishUseCase) {
				duc.EXPECT().CreateDish(gomock.Any(), domain.Dish{
					RestaurantBrandID: 1,
					Name:              "Борщ",
					Description:       "Вкусный",
					Price:             500000000,
				}, gomock.Any(), "dish-key-1").Return(domain.Dish{
					ID:                10,
					RestaurantBrandID: 1,
					Name:              "Борщ",
					Price:             500000000,
				}, nil)
			},
			expectedStatus: codes.OK,
		},
		{
			name: "Ошибка: ресторан не найден",
			req:  &pb.CreateDishRequest{RestaurantBrandId: 999},
			mockInit: func(duc *ucMocks.MockDishUseCase) {
				duc.EXPECT().CreateDish(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(domain.Dish{}, domain.ErrRestaurantNotFound)
			},
			expectedStatus: codes.NotFound,
		},
		{
			name: "Ошибка: некорректные входные данные",
			req:  &pb.CreateDishRequest{Name: ""},
			mockInit: func(duc *ucMocks.MockDishUseCase) {
				duc.EXPECT().CreateDish(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(domain.Dish{}, domain.ErrInvalidInput)
			},
			expectedStatus: codes.InvalidArgument,
		},
		{
			name: "Ошибка при загрузке изображения",
			req:  &pb.CreateDishRequest{Name: "Dish"},
			mockInit: func(duc *ucMocks.MockDishUseCase) {
				duc.EXPECT().CreateDish(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(domain.Dish{}, errors.New("s3 upload failed"))
			},
			expectedStatus: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			duc := ucMocks.NewMockDishUseCase(ctrl)
			h := NewRestaurantHandler(nil, nil, duc)

			tt.mockInit(duc)

			resp, err := h.CreateDish(context.Background(), tt.req)

			if tt.expectedStatus == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.req.Name, resp.Name)
			} else {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedStatus, st.Code())
			}
		})
	}
}

func TestRestaurantHandler_UpdateDish(t *testing.T) {
	type mockInit func(duc *ucMocks.MockDishUseCase)

	newName := "Обновленное блюдо"
	newDesc := "Новое описание"
	newPrice := int64(150000000)

	tests := []struct {
		name           string
		req            *pb.UpdateDishRequest
		mockInit       mockInit
		expectedStatus codes.Code
	}{
		{
			name: "Успешное обновление всех полей",
			req: &pb.UpdateDishRequest{
				Id:             1,
				Name:           &newName,
				Description:    &newDesc,
				Price:          &newPrice,
				ImageData:      []byte("fake-image"),
				IdempotencyKey: "idem-1",
			},
			mockInit: func(duc *ucMocks.MockDishUseCase) {
				duc.EXPECT().UpdateDish(gomock.Any(), domain.Dish{
					ID:          1,
					Name:        newName,
					Description: newDesc,
					Price:       newPrice,
				}, []byte("fake-image"), "idem-1").
					Return(domain.Dish{ID: 1, Name: newName, Price: newPrice}, nil)
			},
			expectedStatus: codes.OK,
		},
		{
			name: "Успешное частичное обновление",
			req: &pb.UpdateDishRequest{
				Id: 1,
			},
			mockInit: func(duc *ucMocks.MockDishUseCase) {
				duc.EXPECT().UpdateDish(gomock.Any(), domain.Dish{ID: 1}, gomock.Any(), gomock.Any()).
					Return(domain.Dish{ID: 1, Name: "Old Name"}, nil)
			},
			expectedStatus: codes.OK,
		},
		{
			name: "Ошибка: блюдо не найдено",
			req:  &pb.UpdateDishRequest{Id: 404},
			mockInit: func(duc *ucMocks.MockDishUseCase) {
				duc.EXPECT().UpdateDish(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(domain.Dish{}, domain.ErrDishNotFound)
			},
			expectedStatus: codes.NotFound,
		},
		{
			name: "Внутренняя ошибка сервера",
			req:  &pb.UpdateDishRequest{Id: 500},
			mockInit: func(duc *ucMocks.MockDishUseCase) {
				duc.EXPECT().UpdateDish(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(domain.Dish{}, errors.New("db error"))
			},
			expectedStatus: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			duc := ucMocks.NewMockDishUseCase(ctrl)
			h := NewRestaurantHandler(nil, nil, duc)

			tt.mockInit(duc)

			resp, err := h.UpdateDish(context.Background(), tt.req)

			if tt.expectedStatus == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			} else {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedStatus, st.Code())
			}
		})
	}
}

func TestRestaurantHandler_DeleteDish(t *testing.T) {
	type mockInit func(duc *ucMocks.MockDishUseCase)

	tests := []struct {
		name           string
		req            *pb.DeleteDishRequest
		mockInit       mockInit
		expectedStatus codes.Code
	}{
		{
			name: "Успешное удаление блюда",
			req:  &pb.DeleteDishRequest{Id: 1},
			mockInit: func(duc *ucMocks.MockDishUseCase) {
				duc.EXPECT().DeleteDish(gomock.Any(), int64(1)).Return(nil)
			},
			expectedStatus: codes.OK,
		},
		{
			name: "Ошибка: блюдо не существует",
			req:  &pb.DeleteDishRequest{Id: 404},
			mockInit: func(duc *ucMocks.MockDishUseCase) {
				duc.EXPECT().DeleteDish(gomock.Any(), int64(404)).Return(domain.ErrDishNotFound)
			},
			expectedStatus: codes.NotFound,
		},
		{
			name: "Внутренняя ошибка сервера при удалении",
			req:  &pb.DeleteDishRequest{Id: 1},
			mockInit: func(duc *ucMocks.MockDishUseCase) {
				duc.EXPECT().DeleteDish(gomock.Any(), int64(1)).Return(errors.New("db crash"))
			},
			expectedStatus: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			duc := ucMocks.NewMockDishUseCase(ctrl)
			h := NewRestaurantHandler(nil, nil, duc)

			tt.mockInit(duc)

			resp, err := h.DeleteDish(context.Background(), tt.req)

			if tt.expectedStatus == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			} else {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedStatus, st.Code())
			}
		})
	}
}

func TestRestaurantHandler_GetCategories(t *testing.T) {
	type mockInit func(cuc *ucMocks.MockCategoryUseCase)

	tests := []struct {
		name           string
		mockInit       mockInit
		expectedStatus codes.Code
	}{
		{
			name: "Успешное получение списка категорий",
			mockInit: func(cuc *ucMocks.MockCategoryUseCase) {
				cuc.EXPECT().GetAllCategories(gomock.Any()).
					Return([]domain.Category{{ID: 1, Name: "Бургеры", Emoji: "🍔"}}, nil)
			},
			expectedStatus: codes.OK,
		},
		{
			name: "Ошибка UseCase при получении категорий",
			mockInit: func(cuc *ucMocks.MockCategoryUseCase) {
				cuc.EXPECT().GetAllCategories(gomock.Any()).
					Return(nil, errors.New("db fail"))
			},
			expectedStatus: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cuc := ucMocks.NewMockCategoryUseCase(ctrl)
			h := NewRestaurantHandler(cuc, nil, nil)

			tt.mockInit(cuc)

			resp, err := h.GetCategories(context.Background(), &emptypb.Empty{})

			if tt.expectedStatus == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Len(t, resp.Categories, 1)
			} else {
				st, _ := status.FromError(err)
				assert.Equal(t, tt.expectedStatus, st.Code())
			}
		})
	}
}
