package restaurantclient

import (
	"context"
	"errors"
	"testing"

	pbRestaurant "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/restaurant"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/restaurant/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestRestaurantClient_GetRestaurantBrandsList(t *testing.T) {
	type mockInit func(m *mocks.MockRestaurantServiceClient)

	tests := []struct {
		name      string
		limit     int32
		offset    int32
		mockInit  mockInit
		wantErr   error
		wantCount int
	}{
		{
			name:   "Успешное получение списка брендов",
			limit:  10,
			offset: 0,
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().GetRestaurantBrandsList(gomock.Any(), &pbRestaurant.GetRestaurantBrandsListRequest{
					Limit: 10, Offset: 0,
				}).Return(&pbRestaurant.GetRestaurantBrandsListResponse{
					RestaurantBrands: []*pbRestaurant.RestaurantBrand{{Id: 1}, {Id: 2}},
				}, nil)
			},
			wantErr:   nil,
			wantCount: 2,
		},
		{
			name:   "Ошибка gRPC возвращает внутреннюю ошибку",
			limit:  10,
			offset: 0,
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().GetRestaurantBrandsList(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("rpc error"))
			},
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockRestaurantServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewRestaurantClient(mockSvc)
			res, err := client.GetRestaurantBrandsList(context.Background(), tt.limit, tt.offset)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.Len(t, res, tt.wantCount)
			}
		})
	}
}

func TestRestaurantClient_GetRestaurantBrandByID(t *testing.T) {
	type mockInit func(m *mocks.MockRestaurantServiceClient)

	tests := []struct {
		name     string
		id       int64
		mockInit mockInit
		wantErr  error
	}{
		{
			name: "Успешное получение бренда по ID",
			id:   1,
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().GetRestaurantBrandByID(gomock.Any(), &pbRestaurant.GetRestaurantBrandByIDRequest{Id: 1}).
					Return(&pbRestaurant.GetRestaurantBrandByIDResponse{
						RestaurantBrand: &pbRestaurant.RestaurantBrand{Id: 1, Name: "Test"},
					}, nil)
			},
			wantErr: nil,
		},
		{
			name: "Бренд не найден",
			id:   404,
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().GetRestaurantBrandByID(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			wantErr: ErrNotFound,
		},
		{
			name: "Прочая gRPC ошибка",
			id:   1,
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().GetRestaurantBrandByID(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.Internal, "crash"))
			},
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockRestaurantServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewRestaurantClient(mockSvc)
			res, err := client.GetRestaurantBrandByID(context.Background(), tt.id)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, res.ID)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, tt.id, res.ID)
			}
		})
	}
}

func TestRestaurantClient_GetDishesByRestaurantBrandID(t *testing.T) {
	type mockInit func(m *mocks.MockRestaurantServiceClient)

	tests := []struct {
		name     string
		brandID  int64
		mockInit mockInit
		wantErr  error
	}{
		{
			name:    "Успешное получение блюд",
			brandID: 10,
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().GetDishesByRestaurantBrandID(gomock.Any(), &pbRestaurant.GetDishesByRestaurantBrandIDRequest{
					RestaurantBrandId: 10, Limit: 20, Offset: 0,
				}).Return(&pbRestaurant.GetDishesByRestaurantBrandIDResponse{
					Dishes: []*pbRestaurant.Dish{{Id: 101}, {Id: 102}},
				}, nil)
			},
			wantErr: nil,
		},
		{
			name:    "Ресторан или блюда не найдены",
			brandID: 10,
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().GetDishesByRestaurantBrandID(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			wantErr: ErrNotFound,
		},
		{
			name:    "Ошибка соединения",
			brandID: 10,
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().GetDishesByRestaurantBrandID(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.Unavailable, "unavailable"))
			},
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockRestaurantServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewRestaurantClient(mockSvc)
			res, err := client.GetDishesByRestaurantBrandID(context.Background(), tt.brandID, 20, 0)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, res)
			}
		})
	}
}

func TestRestaurantClient_GetDishesByIDs(t *testing.T) {
	type mockInit func(m *mocks.MockRestaurantServiceClient)

	tests := []struct {
		name     string
		ids      []int64
		mockInit mockInit
		wantErr  error
	}{
		{
			name: "Успешное получение блюд по списку ID",
			ids:  []int64{1, 2},
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().GetDishesByIDs(gomock.Any(), &pbRestaurant.GetDishesByIDsRequest{
					DishIds: []int64{1, 2},
				}).Return(&pbRestaurant.GetDishesByIDsResponse{
					Dishes: []*pbRestaurant.Dish{{Id: 1}, {Id: 2}},
				}, nil)
			},
			wantErr: nil,
		},
		{
			name: "Внутренняя ошибка при сбое gRPC",
			ids:  []int64{1},
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().GetDishesByIDs(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("rpc error"))
			},
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockRestaurantServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewRestaurantClient(mockSvc)
			res, err := client.GetDishesByIDs(context.Background(), tt.ids)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.Len(t, res, len(tt.ids))
			}
		})
	}
}

func TestRestaurantClient_GetRestaurantLogos(t *testing.T) {
	type mockInit func(m *mocks.MockRestaurantServiceClient)

	tests := []struct {
		name     string
		ids      []int64
		mockInit mockInit
		wantErr  error
		wantMap  map[int64]string
	}{
		{
			name: "Успешное формирование мапы логотипов",
			ids:  []int64{10, 20},
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().GetRestaurantBrandsByIDs(gomock.Any(), &pbRestaurant.GetRestaurantBrandsByIDsRequest{
					BrandIds: []int64{10, 20},
				}).Return(&pbRestaurant.GetRestaurantBrandsByIDsResponse{
					RestaurantBrands: []*pbRestaurant.RestaurantBrand{
						{Id: 10, LogoUrl: "logo10.png"},
						{Id: 20, LogoUrl: "logo20.png"},
					},
				}, nil)
			},
			wantErr: nil,
			wantMap: map[int64]string{10: "logo10.png", 20: "logo20.png"},
		},
		{
			name: "Ошибка gRPC при получении логотипов",
			ids:  []int64{10},
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().GetRestaurantBrandsByIDs(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("fail"))
			},
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockRestaurantServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewRestaurantClient(mockSvc)
			res, err := client.GetRestaurantLogos(context.Background(), tt.ids)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantMap, res)
			}
		})
	}
}

func TestRestaurantClient_GetDishByID(t *testing.T) {
	type mockInit func(m *mocks.MockRestaurantServiceClient)

	tests := []struct {
		name     string
		id       int64
		mockInit mockInit
		wantErr  error
	}{
		{
			name: "Успешное получение одного блюда",
			id:   1,
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().GetDishesByIDs(gomock.Any(), &pbRestaurant.GetDishesByIDsRequest{
					DishIds: []int64{1},
				}).Return(&pbRestaurant.GetDishesByIDsResponse{
					Dishes: []*pbRestaurant.Dish{{Id: 1, Name: "Burger"}},
				}, nil)
			},
			wantErr: nil,
		},
		{
			name: "Блюдо не найдено (пустой список от сервиса)",
			id:   404,
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().GetDishesByIDs(gomock.Any(), gomock.Any()).
					Return(&pbRestaurant.GetDishesByIDsResponse{Dishes: []*pbRestaurant.Dish{}}, nil)
			},
			wantErr: ErrNotFound,
		},
		{
			name: "Ошибка gRPC при поиске блюда",
			id:   1,
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().GetDishesByIDs(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("conn lost"))
			},
			wantErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockRestaurantServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewRestaurantClient(mockSvc)
			res, err := client.GetDishByID(context.Background(), tt.id)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, res.ID)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, tt.id, res.ID)
			}
		})
	}
}

func TestRestaurantClient_CreateRestaurantBrand(t *testing.T) {
	type mockInit func(m *mocks.MockRestaurantServiceClient)

	tests := []struct {
		name     string
		mockInit mockInit
		wantErr  bool
	}{
		{
			name: "Успешное создание бренда",
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().CreateRestaurantBrand(gomock.Any(), &pbRestaurant.CreateBrandRequest{
					OwnerId:        1,
					Name:           "KFC",
					Description:    "Chicken",
					LogoData:       []byte("logo"),
					IdempotencyKey: "key-1",
				}).Return(&pbRestaurant.RestaurantBrand{Id: 10, Name: "KFC"}, nil)
			},
			wantErr: false,
		},
		{
			name: "Ошибка gRPC при создании",
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().CreateRestaurantBrand(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("rpc error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockRestaurantServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewRestaurantClient(mockSvc)
			res, err := client.CreateRestaurantBrand(context.Background(), 1, "KFC", "Chicken", []byte("logo"), "key-1")

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, RestaurantBrand{}, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
			}
		})
	}
}

func TestRestaurantClient_UpdateRestaurantBrand(t *testing.T) {
	type mockInit func(m *mocks.MockRestaurantServiceClient)

	name := "Updated KFC"
	desc := "New chicken"
	tier := int32(3)

	tests := []struct {
		name     string
		namePtr  *string
		descPtr  *string
		tierPtr  *int32
		mockInit mockInit
		wantErr  bool
	}{
		{
			name:    "Успешное полное обновление",
			namePtr: &name,
			descPtr: &desc,
			tierPtr: &tier,
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().UpdateRestaurantBrand(gomock.Any(), &pbRestaurant.UpdateBrandRequest{
					Id:             10,
					Name:           &name,
					Description:    &desc,
					PromotionTier:  &tier,
					LogoData:       []byte("new-logo"),
					IdempotencyKey: "upd-key",
				}).Return(&pbRestaurant.RestaurantBrand{Id: 10}, nil)
			},
			wantErr: false,
		},
		{
			name:    "Частичное обновление (только лого)",
			namePtr: nil,
			descPtr: nil,
			tierPtr: nil,
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().UpdateRestaurantBrand(gomock.Any(), &pbRestaurant.UpdateBrandRequest{
					Id:             10,
					LogoData:       []byte("new-logo"),
					IdempotencyKey: "upd-key",
				}).Return(&pbRestaurant.RestaurantBrand{Id: 10}, nil)
			},
			wantErr: false,
		},
		{
			name:    "Ошибка при обновлении",
			namePtr: nil,
			descPtr: nil,
			tierPtr: nil,
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().UpdateRestaurantBrand(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("fail"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockRestaurantServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewRestaurantClient(mockSvc)
			res, err := client.UpdateRestaurantBrand(context.Background(), 10, tt.namePtr, tt.descPtr, []byte("new-logo"), tt.tierPtr, "upd-key")

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, RestaurantBrand{}, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
			}
		})
	}
}

func TestRestaurantClient_DeleteRestaurantBrand(t *testing.T) {
	type mockInit func(m *mocks.MockRestaurantServiceClient)

	tests := []struct {
		name     string
		mockInit mockInit
		wantErr  bool
	}{
		{
			name: "Успешное удаление",
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().DeleteRestaurantBrand(gomock.Any(), &pbRestaurant.DeleteBrandRequest{Id: 10}).
					Return(&emptypb.Empty{}, nil)
			},
			wantErr: false,
		},
		{
			name: "Ошибка при удалении",
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().DeleteRestaurantBrand(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("delete error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockRestaurantServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewRestaurantClient(mockSvc)
			err := client.DeleteRestaurantBrand(context.Background(), 10)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRestaurantClient_CreateDish(t *testing.T) {
	type mockInit func(m *mocks.MockRestaurantServiceClient)

	tests := []struct {
		name     string
		mockInit mockInit
		wantErr  bool
	}{
		{
			name: "Успешное создание блюда",
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().CreateDish(gomock.Any(), &pbRestaurant.CreateDishRequest{
					RestaurantBrandId: 1,
					Name:              "Бургер",
					Description:       "Вкусный",
					Price:             500,
					ImageData:         []byte("data"),
					IdempotencyKey:    "key-1",
				}).Return(&pbRestaurant.Dish{Id: 10, Name: "Бургер"}, nil)
			},
			wantErr: false,
		},
		{
			name: "Ошибка gRPC при создании блюда",
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().CreateDish(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("rpc error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockRestaurantServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewRestaurantClient(mockSvc)
			res, err := client.CreateDish(context.Background(), 1, "Бургер", "Вкусный", 500, []byte("data"), "key-1")

			if tt.wantErr {
				assert.Error(t, err)
				assert.Zero(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
			}
		})
	}
}

func TestRestaurantClient_UpdateDish(t *testing.T) {
	type mockInit func(m *mocks.MockRestaurantServiceClient)

	name := "Новое имя"
	desc := "Новое описание"
	price := int64(600)

	tests := []struct {
		name     string
		namePtr  *string
		descPtr  *string
		pricePtr *int64
		mockInit mockInit
		wantErr  bool
	}{
		{
			name:     "Успешное полное обновление блюда",
			namePtr:  &name,
			descPtr:  &desc,
			pricePtr: &price,
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().UpdateDish(gomock.Any(), &pbRestaurant.UpdateDishRequest{
					Id:             10,
					Name:           &name,
					Description:    &desc,
					Price:          &price,
					ImageData:      []byte("new-img"),
					IdempotencyKey: "upd-key",
				}).Return(&pbRestaurant.Dish{Id: 10}, nil)
			},
			wantErr: false,
		},
		{
			name:     "Частичное обновление блюда (только цена)",
			namePtr:  nil,
			descPtr:  nil,
			pricePtr: &price,
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().UpdateDish(gomock.Any(), &pbRestaurant.UpdateDishRequest{
					Id:             10,
					Price:          &price,
					ImageData:      []byte("new-img"),
					IdempotencyKey: "upd-key",
				}).Return(&pbRestaurant.Dish{Id: 10}, nil)
			},
			wantErr: false,
		},
		{
			name: "Ошибка gRPC при обновлении блюда",
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().UpdateDish(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("fail"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockRestaurantServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewRestaurantClient(mockSvc)
			res, err := client.UpdateDish(context.Background(), 10, tt.namePtr, tt.descPtr, tt.pricePtr, []byte("new-img"), "upd-key")

			if tt.wantErr {
				assert.Error(t, err)
				assert.Zero(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
			}
		})
	}
}

func TestRestaurantClient_DeleteDish(t *testing.T) {
	type mockInit func(m *mocks.MockRestaurantServiceClient)

	tests := []struct {
		name     string
		mockInit mockInit
		wantErr  bool
	}{
		{
			name: "Успешное удаление блюда",
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().DeleteDish(gomock.Any(), &pbRestaurant.DeleteDishRequest{Id: 10}).
					Return(&emptypb.Empty{}, nil)
			},
			wantErr: false,
		},
		{
			name: "Ошибка gRPC при удалении блюда",
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().DeleteDish(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("delete error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockRestaurantServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewRestaurantClient(mockSvc)
			err := client.DeleteDish(context.Background(), 10)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRestaurantClient_SearchRestaurantBrands(t *testing.T) {
	type mockInit func(m *mocks.MockRestaurantServiceClient)

	query := "Бургер King"

	tests := []struct {
		name     string
		mockInit mockInit
		wantErr  error
	}{
		{
			name: "Успешный поиск ресторанов",
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().SearchRestaurantBrands(gomock.Any(), &pbRestaurant.SearchRestaurantBrandsRequest{
					Query:  query,
					Limit:  10,
					Offset: 0,
				}).Return(&pbRestaurant.GetRestaurantBrandsListResponse{
					RestaurantBrands: []*pbRestaurant.RestaurantBrand{{Id: 100}},
				}, nil)
			},
			wantErr: nil,
		},
		{
			name: "Ошибка gRPC при поиске",
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().SearchRestaurantBrands(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("search error"))
			},
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockRestaurantServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewRestaurantClient(mockSvc)
			res, err := client.SearchRestaurantBrands(context.Background(), query, 10, 0)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Len(t, res, 1)
			}
		})
	}
}

func TestRestaurantClient_SearchDishes(t *testing.T) {
	type mockInit func(m *mocks.MockRestaurantServiceClient)

	tests := []struct {
		name     string
		query    string
		limit    int32
		mockInit mockInit
		wantErr  error
	}{
		{
			name:  "Успешный глобальный поиск блюд",
			query: "бургер",
			limit: 10,
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().SearchDishes(gomock.Any(), &pbRestaurant.SearchDishesRequest{
					Query: "бургер",
					Limit: 10,
				}).Return(&pbRestaurant.GetDishesByIDsResponse{
					Dishes: []*pbRestaurant.Dish{{Id: 1, Name: "Бургер"}},
				}, nil)
			},
			wantErr: nil,
		},
		{
			name:  "Ошибка gRPC при глобальном поиске",
			query: "pizza",
			limit: 5,
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().SearchDishes(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("rpc error"))
			},
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockRestaurantServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewRestaurantClient(mockSvc)
			res, err := client.SearchDishes(context.Background(), tt.query, tt.limit)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
			}
		})
	}
}

func TestRestaurantClient_SearchDishesByBrand(t *testing.T) {
	type mockInit func(m *mocks.MockRestaurantServiceClient)

	tests := []struct {
		name     string
		brandID  int64
		query    string
		limit    int32
		mockInit mockInit
		wantErr  error
	}{
		{
			name:    "Успешный поиск блюд внутри бренда",
			brandID: 1,
			query:   "cola",
			limit:   20,
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().SearchDishesByBrand(gomock.Any(), &pbRestaurant.SearchDishesByBrandRequest{
					RestaurantBrandId: 1,
					Query:             "cola",
					Limit:             20,
				}).Return(&pbRestaurant.GetDishesByRestaurantBrandIDResponse{
					Dishes: []*pbRestaurant.Dish{{Id: 10, Name: "Cola"}},
				}, nil)
			},
			wantErr: nil,
		},
		{
			name:    "Ошибка gRPC при поиске по бренду",
			brandID: 2,
			query:   "fries",
			limit:   10,
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().SearchDishesByBrand(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("fail"))
			},
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockRestaurantServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewRestaurantClient(mockSvc)
			res, err := client.SearchDishesByBrand(context.Background(), tt.brandID, tt.query, tt.limit)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
			}
		})
	}
}

func TestRestaurantClient_GetRestaurantBrandsByCategoryName(t *testing.T) {
	type mockInit func(m *mocks.MockRestaurantServiceClient)

	catName := "Пицца и Суши"

	tests := []struct {
		name     string
		mockInit mockInit
		wantErr  bool
	}{
		{
			name: "Успешное получение по имени категории",
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().GetRestaurantBrandsByCategory(gomock.Any(), &pbRestaurant.GetRestaurantBrandsByCategoryRequest{
					CategoryName: catName,
					Limit:        10,
					Offset:       0,
				}).Return(&pbRestaurant.GetRestaurantBrandsListResponse{
					RestaurantBrands: []*pbRestaurant.RestaurantBrand{{Id: 1}},
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "Ошибка gRPC при получении по имени",
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().GetRestaurantBrandsByCategory(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("fail"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockRestaurantServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewRestaurantClient(mockSvc)

			res, err := client.GetRestaurantBrandsByCategoryName(context.Background(), catName, 10, 0)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, res)
				assert.Equal(t, int64(1), res[0].ID)
			}
		})
	}
}
