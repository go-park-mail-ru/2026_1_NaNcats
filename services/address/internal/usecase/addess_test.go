package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/address/internal/domain"
	repoMocks "github.com/go-park-mail-ru/2026_1_NaNcats/services/address/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAddressUseCase_AddAddress(t *testing.T) {
	type mockInit func(r *repoMocks.MockAddressRepository)

	userID := int64(1)
	addr := domain.Address{Label: "Дом", Apartment: "42"}
	idemKey := "idem-123"

	tests := []struct {
		name     string
		mockInit mockInit
		want     string
		wantErr  error
	}{
		{
			name: "Успешное добавление адреса",
			mockInit: func(r *repoMocks.MockAddressRepository) {
				r.EXPECT().CreateAddress(gomock.Any(), userID, addr, idemKey).
					Return("uuid-addr", nil)
			},
			want:    "uuid-addr",
			wantErr: nil,
		},
		{
			name: "Ошибка при создании в репозитории",
			mockInit: func(r *repoMocks.MockAddressRepository) {
				r.EXPECT().CreateAddress(gomock.Any(), userID, addr, idemKey).
					Return("", errors.New("db error"))
			},
			want:    "",
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := repoMocks.NewMockAddressRepository(ctrl)
			tt.mockInit(repo)

			uc := NewAddressUseCase(repo)
			got, err := uc.AddAddress(context.Background(), userID, addr, idemKey)

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAddressUseCase_GetMyAddresses(t *testing.T) {
	type mockInit func(r *repoMocks.MockAddressRepository)

	userID := int64(1)
	addressList := []domain.Address{
		{PublicID: "1", Label: "Дом"},
		{PublicID: "2", Label: "Работа"},
	}

	tests := []struct {
		name     string
		mockInit mockInit
		want     []domain.Address
		wantErr  error
	}{
		{
			name: "Успешное получение списка адресов",
			mockInit: func(r *repoMocks.MockAddressRepository) {
				r.EXPECT().GetAddressesByUserID(gomock.Any(), userID).
					Return(addressList, nil).Times(2)
			},
			want:    addressList,
			wantErr: nil,
		},
		{
			name: "Успешное получение пустого списка",
			mockInit: func(r *repoMocks.MockAddressRepository) {
				r.EXPECT().GetAddressesByUserID(gomock.Any(), userID).
					Return([]domain.Address{}, nil).Times(2)
			},
			want:    []domain.Address{},
			wantErr: nil,
		},
		{
			name: "Ошибка репозитория",
			mockInit: func(r *repoMocks.MockAddressRepository) {
				r.EXPECT().GetAddressesByUserID(gomock.Any(), userID).
					Return(nil, errors.New("failed to fetch")).Times(1)
			},
			want:    nil,
			wantErr: errors.New("failed to fetch"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := repoMocks.NewMockAddressRepository(ctrl)
			tt.mockInit(repo)

			uc := NewAddressUseCase(repo)
			got, err := uc.GetMyAddresses(context.Background(), userID)

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAddressUseCase_DeleteAddress(t *testing.T) {
	type mockInit func(r *repoMocks.MockAddressRepository)

	userID := int64(1)
	publicID := "addr-uuid"
	idemKey := "idem-delete"

	tests := []struct {
		name     string
		mockInit mockInit
		wantErr  error
	}{
		{
			name: "Успешное удаление",
			mockInit: func(r *repoMocks.MockAddressRepository) {
				r.EXPECT().DeleteAddress(gomock.Any(), userID, publicID).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "Адрес не найден",
			mockInit: func(r *repoMocks.MockAddressRepository) {
				r.EXPECT().DeleteAddress(gomock.Any(), userID, publicID).
					Return(domain.ErrAddressNotFound)
			},
			wantErr: domain.ErrAddressNotFound,
		},
		{
			name: "Системная ошибка репозитория",
			mockInit: func(r *repoMocks.MockAddressRepository) {
				r.EXPECT().DeleteAddress(gomock.Any(), userID, publicID).
					Return(errors.New("internal"))
			},
			wantErr: errors.New("internal"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := repoMocks.NewMockAddressRepository(ctrl)
			tt.mockInit(repo)

			uc := NewAddressUseCase(repo)
			err := uc.DeleteAddress(context.Background(), userID, publicID, idemKey)

			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestAddressUseCase_UpdateAddress(t *testing.T) {
	type mockInit func(r *repoMocks.MockAddressRepository)

	userID := int64(1)
	addr := domain.Address{
		PublicID:  "addr-uuid-123",
		Apartment: "101",
		Label:     "Дом",
	}
	idemKey := "idem-update-456"

	tests := []struct {
		name     string
		mockInit mockInit
		wantErr  error
	}{
		{
			name: "Успешное обновление адреса",
			mockInit: func(r *repoMocks.MockAddressRepository) {
				r.EXPECT().UpdateAddress(gomock.Any(), userID, addr).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "Ошибка: адрес не найден",
			mockInit: func(r *repoMocks.MockAddressRepository) {
				r.EXPECT().UpdateAddress(gomock.Any(), userID, addr).
					Return(domain.ErrAddressNotFound)
			},
			wantErr: domain.ErrAddressNotFound,
		},
		{
			name: "Системная ошибка при обновлении",
			mockInit: func(r *repoMocks.MockAddressRepository) {
				r.EXPECT().UpdateAddress(gomock.Any(), userID, addr).
					Return(errors.New("db failure"))
			},
			wantErr: errors.New("db failure"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := repoMocks.NewMockAddressRepository(ctrl)
			tt.mockInit(repo)

			uc := NewAddressUseCase(repo)
			err := uc.UpdateAddress(context.Background(), userID, addr, idemKey)

			assert.Equal(t, tt.wantErr, err)
		})
	}
}
