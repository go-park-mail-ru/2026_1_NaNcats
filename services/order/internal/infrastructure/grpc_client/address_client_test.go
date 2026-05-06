package grpc_client

import (
	"context"
	"errors"
	"testing"

	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/address"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/address/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/emptypb"
)

//go:generate mockgen -destination=../../../../../shared/proto/address/mocks/address_client_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/address AddressServiceClient

func TestAddressClient_CheckAddressExists(t *testing.T) {
	type mockBehavior func(mock *mocks.MockAddressServiceClient, userID int64, addressPublicID string)

	tests := []struct {
		name            string
		userID          int64
		addressPublicID string
		mockBehavior    mockBehavior
		expectedErr     error
	}{
		{
			name:            "Успешная проверка адреса",
			userID:          1,
			addressPublicID: "addr-123",
			mockBehavior: func(m *mocks.MockAddressServiceClient, userID int64, addressPublicID string) {
				req := &pb.CheckAddressExistsRequest{
					UserId:          userID,
					AddressPublicId: addressPublicID,
				}
				m.EXPECT().CheckAddressExists(gomock.Any(), req).Return(&emptypb.Empty{}, nil)
			},
			expectedErr: nil,
		},
		{
			name:            "Ошибка: адрес не найден",
			userID:          2,
			addressPublicID: "addr-404",
			mockBehavior: func(m *mocks.MockAddressServiceClient, userID int64, addressPublicID string) {
				req := &pb.CheckAddressExistsRequest{
					UserId:          userID,
					AddressPublicId: addressPublicID,
				}
				m.EXPECT().CheckAddressExists(gomock.Any(), req).Return(nil, errors.New("not found"))
			},
			expectedErr: errors.New("not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockAddressServiceClient(ctrl)
			tt.mockBehavior(mockClient, tt.userID, tt.addressPublicID)

			client := NewAddressClient(mockClient)
			err := client.CheckAddressExists(context.Background(), tt.userID, tt.addressPublicID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
