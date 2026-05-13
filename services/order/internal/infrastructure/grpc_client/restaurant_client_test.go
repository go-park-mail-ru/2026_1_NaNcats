package grpc_client

import (
	"context"
	"errors"
	"testing"

	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/restaurant"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/restaurant/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -destination=../../../../../shared/proto/restaurant/mocks/restaurant_client_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/restaurant RestaurantServiceClient

func TestRestaurantClient_GetRestaurantName(t *testing.T) {
	type mockBehavior func(mock *mocks.MockRestaurantServiceClient, branchID int64)

	tests := []struct {
		name         string
		branchID     int64
		mockBehavior mockBehavior
		expectedName string
		expectedErr  error
	}{
		{
			name:     "Успешное получение имени",
			branchID: 100,
			mockBehavior: func(m *mocks.MockRestaurantServiceClient, branchID int64) {
				req := &pb.GetRestaurantBrandByIDRequest{Id: branchID}
				resp := &pb.GetRestaurantBrandByIDResponse{
					RestaurantBrand: &pb.RestaurantBrand{Name: "Вкусно и Точка"},
				}
				m.EXPECT().GetRestaurantBrandByID(gomock.Any(), req).Return(resp, nil)
			},
			expectedName: "Вкусно и Точка",
			expectedErr:  nil,
		},
		{
			name:     "Ошибка: бренд не найден",
			branchID: 404,
			mockBehavior: func(m *mocks.MockRestaurantServiceClient, branchID int64) {
				req := &pb.GetRestaurantBrandByIDRequest{Id: branchID}
				m.EXPECT().GetRestaurantBrandByID(gomock.Any(), req).Return(nil, errors.New("not found"))
			},
			expectedName: "",
			expectedErr:  errors.New("not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockRestaurantServiceClient(ctrl)
			tt.mockBehavior(mockClient, tt.branchID)

			client := NewRestaurantClient(mockClient)
			name, err := client.GetRestaurantName(context.Background(), tt.branchID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedName, name)
			}
		})
	}
}
