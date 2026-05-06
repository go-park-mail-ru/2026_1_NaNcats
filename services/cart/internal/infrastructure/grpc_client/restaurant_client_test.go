package grpc_client

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/domain"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/restaurant"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/restaurant/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRestaurantClient_GetDishesByIDs(t *testing.T) {
	type mockInit func(m *mocks.MockRestaurantServiceClient)

	tests := []struct {
		name           string
		dishIDs        []int64
		mockInit       mockInit
		expectedDishes []domain.Dish
		expectedErr    error
	}{
		{
			name:    "Успешное получение списка блюд",
			dishIDs: []int64{1, 2},
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().GetDishesByIDs(gomock.Any(), &pb.GetDishesByIDsRequest{
					DishIds: []int64{1, 2},
				}).Return(&pb.GetDishesByIDsResponse{
					Dishes: []*pb.Dish{
						{
							Id:                1,
							Name:              "Борщ",
							Price:             500,
							ImageUrl:          "borsch.jpg",
							RestaurantBrandId: 10,
						},
						{
							Id:                2,
							Name:              "Плов",
							Price:             600,
							ImageUrl:          "plov.jpg",
							RestaurantBrandId: 10,
						},
					},
				}, nil)
			},
			expectedDishes: []domain.Dish{
				{ID: 1, Name: "Борщ", Price: 500, ImageURL: "borsch.jpg", RestaurantBrandID: 10},
				{ID: 2, Name: "Плов", Price: 600, ImageURL: "plov.jpg", RestaurantBrandID: 10},
			},
			expectedErr: nil,
		},
		{
			name:    "Успешное получение пустого списка",
			dishIDs: []int64{999},
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().GetDishesByIDs(gomock.Any(), &pb.GetDishesByIDsRequest{
					DishIds: []int64{999},
				}).Return(&pb.GetDishesByIDsResponse{
					Dishes: []*pb.Dish{},
				}, nil)
			},
			expectedDishes: []domain.Dish{},
			expectedErr:    nil,
		},
		{
			name:    "Ошибка gRPC соединения",
			dishIDs: []int64{1},
			mockInit: func(m *mocks.MockRestaurantServiceClient) {
				m.EXPECT().GetDishesByIDs(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("grpc connection error"))
			},
			expectedDishes: nil,
			expectedErr:    errors.New("grpc connection error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockServiceClient := mocks.NewMockRestaurantServiceClient(ctrl)
			tt.mockInit(mockServiceClient)

			client := NewRestaurantClient(mockServiceClient)

			result, err := client.GetDishesByIDs(context.Background(), tt.dishIDs)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedDishes, result)
				assert.Len(t, result, len(tt.expectedDishes))
			}
		})
	}
}
