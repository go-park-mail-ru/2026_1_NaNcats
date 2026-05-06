package addressclient

import (
	"context"
	"errors"
	"testing"

	pbAddress "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/address"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/address/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAddressClient_AddAddress(t *testing.T) {
	type mockInit func(m *mocks.MockAddressServiceClient)

	userID := int64(1)
	addr := Address{
		Location: Location{
			AddressText: "Moscow",
			Latitude:    55.75,
			Longitude:   37.61,
		},
		Label: "Home",
	}
	idemKey := "idem-123"

	tests := []struct {
		name     string
		mockInit mockInit
		want     string
		wantErr  error
	}{
		{
			name: "Успешное добавление адреса",
			mockInit: func(m *mocks.MockAddressServiceClient) {
				m.EXPECT().AddAddress(gomock.Any(), gomock.Any()).Return(&pbAddress.AddAddressResponse{
					AddressPublicId: "addr-uuid-456",
				}, nil)
			},
			want:    "addr-uuid-456",
			wantErr: nil,
		},
		{
			name: "Ошибка gRPC возвращает ErrInternal",
			mockInit: func(m *mocks.MockAddressServiceClient) {
				m.EXPECT().AddAddress(gomock.Any(), gomock.Any()).Return(nil, errors.New("rpc fail"))
			},
			want:    "",
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockAddressServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewAddressClient(mockSvc)
			res, err := client.AddAddress(context.Background(), userID, addr, idemKey)

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, res)
		})
	}
}

func TestAddressClient_GetMyAddresses(t *testing.T) {
	type mockInit func(m *mocks.MockAddressServiceClient)

	userID := int64(1)

	tests := []struct {
		name     string
		mockInit mockInit
		wantLen  int
		wantErr  error
	}{
		{
			name: "Успешное получение списка адресов",
			mockInit: func(m *mocks.MockAddressServiceClient) {
				m.EXPECT().GetMyAddresses(gomock.Any(), &pbAddress.GetMyAddressesRequest{UserId: userID}).
					Return(&pbAddress.GetMyAddressesResponse{
						Addresses: []*pbAddress.Address{
							{PublicId: "1", Location: &pbAddress.Location{AddressText: "A"}},
							{PublicId: "2", Location: &pbAddress.Location{AddressText: "B"}},
						},
					}, nil)
			},
			wantLen: 2,
			wantErr: nil,
		},
		{
			name: "Успешное получение пустого списка",
			mockInit: func(m *mocks.MockAddressServiceClient) {
				m.EXPECT().GetMyAddresses(gomock.Any(), gomock.Any()).
					Return(&pbAddress.GetMyAddressesResponse{Addresses: []*pbAddress.Address{}}, nil)
			},
			wantLen: 0,
			wantErr: nil,
		},
		{
			name: "Внутренняя ошибка сервиса",
			mockInit: func(m *mocks.MockAddressServiceClient) {
				m.EXPECT().GetMyAddresses(gomock.Any(), gomock.Any()).Return(nil, errors.New("fail"))
			},
			wantLen: 0,
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockAddressServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewAddressClient(mockSvc)
			res, err := client.GetMyAddresses(context.Background(), userID)

			assert.Equal(t, tt.wantErr, err)
			assert.Len(t, res, tt.wantLen)
		})
	}
}

func TestAddressClient_DeleteAddress(t *testing.T) {
	type mockInit func(m *mocks.MockAddressServiceClient)

	userID := int64(1)
	publicID := "uuid-123"
	idemKey := "key-del"

	tests := []struct {
		name     string
		mockInit mockInit
		wantErr  error
	}{
		{
			name: "Успешное удаление",
			mockInit: func(m *mocks.MockAddressServiceClient) {
				m.EXPECT().DeleteAddress(gomock.Any(), &pbAddress.DeleteAddressRequest{
					UserId: userID, AddressPublicId: publicID, IdempotencyKey: idemKey,
				}).Return(nil, nil)
			},
			wantErr: nil,
		},
		{
			name: "Ошибка: адрес не найден",
			mockInit: func(m *mocks.MockAddressServiceClient) {
				m.EXPECT().DeleteAddress(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			wantErr: ErrAddressNotFound,
		},
		{
			name: "Ошибка: системный сбой",
			mockInit: func(m *mocks.MockAddressServiceClient) {
				m.EXPECT().DeleteAddress(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.Internal, "crash"))
			},
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockAddressServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewAddressClient(mockSvc)
			err := client.DeleteAddress(context.Background(), userID, publicID, idemKey)

			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestAddressClient_UpdateAddress(t *testing.T) {
	type mockInit func(m *mocks.MockAddressServiceClient)

	userID := int64(1)
	addr := Address{
		PublicID: "addr-uuid",
		Location: Location{AddressText: "New Moscow"},
	}
	idemKey := "idem-upd-789"

	tests := []struct {
		name     string
		mockInit mockInit
		wantErr  error
	}{
		{
			name: "Успешное обновление адреса",
			mockInit: func(m *mocks.MockAddressServiceClient) {
				m.EXPECT().UpdateAddress(gomock.Any(), gomock.Any()).
					Return(nil, nil)
			},
			wantErr: nil,
		},
		{
			name: "Ошибка: адрес не найден",
			mockInit: func(m *mocks.MockAddressServiceClient) {
				m.EXPECT().UpdateAddress(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			wantErr: ErrAddressNotFound,
		},
		{
			name: "Внутренняя ошибка при обновлении",
			mockInit: func(m *mocks.MockAddressServiceClient) {
				m.EXPECT().UpdateAddress(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("rpc error"))
			},
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockAddressServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewAddressClient(mockSvc)
			err := client.UpdateAddress(context.Background(), userID, addr, idemKey)

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestAddressClient_CheckAddressExists(t *testing.T) {
	type mockInit func(m *mocks.MockAddressServiceClient)

	userID := int64(1)
	publicID := "addr-uuid"

	tests := []struct {
		name     string
		mockInit mockInit
		wantErr  error
	}{
		{
			name: "Адрес существует",
			mockInit: func(m *mocks.MockAddressServiceClient) {
				m.EXPECT().CheckAddressExists(gomock.Any(), &pbAddress.CheckAddressExistsRequest{
					UserId: userID, AddressPublicId: publicID,
				}).Return(nil, nil)
			},
			wantErr: nil,
		},
		{
			name: "Адрес не существует (NotFound)",
			mockInit: func(m *mocks.MockAddressServiceClient) {
				m.EXPECT().CheckAddressExists(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			wantErr: ErrAddressNotFound,
		},
		{
			name: "Ошибка gRPC при проверке",
			mockInit: func(m *mocks.MockAddressServiceClient) {
				m.EXPECT().CheckAddressExists(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.Internal, "crash"))
			},
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockAddressServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewAddressClient(mockSvc)
			err := client.CheckAddressExists(context.Background(), userID, publicID)

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
