package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/address/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/address/internal/usecase/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/grpcutil"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/address"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAddressHandler_AddAddress(t *testing.T) {
	type mockInit func(m *mocks.MockAddressUseCase)

	tests := []struct {
		name         string
		req          *pb.AddAddressRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешное добавление адреса",
			req: &pb.AddAddressRequest{
				UserId: 1,
				Address: &pb.Address{
					Location: &pb.Location{
						AddressText: "Москва, Арбат",
						Latitude:    55.75,
						Longitude:   37.61,
					},
					Label: "Дом",
				},
				IdempotencyKey: "key-123",
			},
			mockInit: func(m *mocks.MockAddressUseCase) {
				m.EXPECT().AddAddress(gomock.Any(), int64(1), gomock.Any(), "key-123").
					Return("addr-uuid-456", nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка: пустое тело адреса",
			req: &pb.AddAddressRequest{
				UserId:  1,
				Address: nil,
			},
			mockInit:     func(m *mocks.MockAddressUseCase) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Ошибка в UseCase",
			req: &pb.AddAddressRequest{
				UserId: 1,
				Address: &pb.Address{
					Location: &pb.Location{},
				},
				IdempotencyKey: "key-err",
			},
			mockInit: func(m *mocks.MockAddressUseCase) {
				m.EXPECT().AddAddress(gomock.Any(), int64(1), gomock.Any(), "key-err").
					Return("", errors.New("internal error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockAddressUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewAddressHandler(mockUC)
			resp, err := handler.AddAddress(context.Background(), tt.req)

			// Нормализуем ошибку через grpcutil, так как хэндлер может возвращать
			// сырой errutil.domainError, который status.FromError не распознает без конвертации
			grpcErr := grpcutil.ToGRPCError(err)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, grpcErr)
				assert.NotNil(t, resp)
				assert.Equal(t, "addr-uuid-456", resp.AddressPublicId)
			} else {
				st, ok := status.FromError(grpcErr)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestAddressHandler_GetMyAddresses(t *testing.T) {
	type mockInit func(m *mocks.MockAddressUseCase)

	tests := []struct {
		name         string
		req          *pb.GetMyAddressesRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешное получение списка адресов",
			req:  &pb.GetMyAddressesRequest{UserId: 1},
			mockInit: func(m *mocks.MockAddressUseCase) {
				m.EXPECT().GetMyAddresses(gomock.Any(), int64(1)).
					Return([]domain.Address{
						{PublicID: "uuid-1", Label: "Дом", Location: domain.Location{AddressText: "Текст 1"}},
						{PublicID: "uuid-2", Label: "Работа", Location: domain.Location{AddressText: "Текст 2"}},
					}, nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Успешное получение пустого списка",
			req:  &pb.GetMyAddressesRequest{UserId: 2},
			mockInit: func(m *mocks.MockAddressUseCase) {
				m.EXPECT().GetMyAddresses(gomock.Any(), int64(2)).
					Return([]domain.Address{}, nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка UseCase при получении",
			req:  &pb.GetMyAddressesRequest{UserId: 1},
			mockInit: func(m *mocks.MockAddressUseCase) {
				m.EXPECT().GetMyAddresses(gomock.Any(), int64(1)).
					Return(nil, errors.New("db fail"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockAddressUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewAddressHandler(mockUC)
			resp, err := handler.GetMyAddresses(context.Background(), tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.IsType(t, []*pb.Address{}, resp.Addresses)
			} else {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestAddressHandler_DeleteAddress(t *testing.T) {
	type mockInit func(m *mocks.MockAddressUseCase)

	tests := []struct {
		name         string
		req          *pb.DeleteAddressRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешное удаление адреса",
			req: &pb.DeleteAddressRequest{
				UserId:          1,
				AddressPublicId: "uuid-123",
				IdempotencyKey:  "idem-del",
			},
			mockInit: func(m *mocks.MockAddressUseCase) {
				m.EXPECT().DeleteAddress(gomock.Any(), int64(1), "uuid-123", "idem-del").
					Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка: адрес не найден",
			req: &pb.DeleteAddressRequest{
				UserId:          1,
				AddressPublicId: "uuid-404",
			},
			mockInit: func(m *mocks.MockAddressUseCase) {
				m.EXPECT().DeleteAddress(gomock.Any(), int64(1), "uuid-404", "").
					Return(domain.ErrAddressNotFound)
			},
			expectedCode: codes.NotFound,
		},
		{
			name: "Системная ошибка в UseCase",
			req:  &pb.DeleteAddressRequest{UserId: 1, AddressPublicId: "uuid-err"},
			mockInit: func(m *mocks.MockAddressUseCase) {
				m.EXPECT().DeleteAddress(gomock.Any(), int64(1), "uuid-err", "").
					Return(errors.New("db crash"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockAddressUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewAddressHandler(mockUC)
			resp, err := handler.DeleteAddress(context.Background(), tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			} else {
				st, _ := status.FromError(err)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestAddressHandler_UpdateAddress(t *testing.T) {
	type mockInit func(m *mocks.MockAddressUseCase)

	tests := []struct {
		name         string
		req          *pb.UpdateAddressRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешное обновление адреса",
			req: &pb.UpdateAddressRequest{
				UserId: 1,
				Address: &pb.Address{
					PublicId: "uuid-123",
					Location: &pb.Location{AddressText: "Новый адрес"},
				},
				IdempotencyKey: "idem-upd",
			},
			mockInit: func(m *mocks.MockAddressUseCase) {
				m.EXPECT().UpdateAddress(gomock.Any(), int64(1), gomock.Any(), "idem-upd").
					Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка: передана пустая структура адреса",
			req: &pb.UpdateAddressRequest{
				UserId:  1,
				Address: nil,
			},
			mockInit:     func(m *mocks.MockAddressUseCase) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Ошибка: адрес не найден при обновлении",
			req: &pb.UpdateAddressRequest{
				UserId: 1,
				Address: &pb.Address{
					PublicId: "uuid-missing",
					Location: &pb.Location{},
				},
			},
			mockInit: func(m *mocks.MockAddressUseCase) {
				m.EXPECT().UpdateAddress(gomock.Any(), int64(1), gomock.Any(), "").
					Return(domain.ErrAddressNotFound)
			},
			expectedCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockAddressUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewAddressHandler(mockUC)
			resp, err := handler.UpdateAddress(context.Background(), tt.req)

			grpcErr := grpcutil.ToGRPCError(err)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, grpcErr)
				assert.NotNil(t, resp)
			} else {
				st, ok := status.FromError(grpcErr)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestAddressHandler_CheckAddressExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUC := mocks.NewMockAddressUseCase(ctrl)
	handler := NewAddressHandler(mockUC)

	t.Run("Успешный вызов заглушки", func(t *testing.T) {
		req := &pb.CheckAddressExistsRequest{
			UserId:          1,
			AddressPublicId: "uuid-123",
		}
		resp, err := handler.CheckAddressExists(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}
