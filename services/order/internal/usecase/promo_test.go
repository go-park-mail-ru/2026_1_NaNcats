package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/repository"
	repomocks "github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func ptrInt(v int) *int       { return &v }
func ptrInt64(v int64) *int64 { return &v }

func futurePromo() domain.Promocode {
	return domain.Promocode{
		ID:        1,
		Code:      "SALE",
		Title:     "Скидка",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
}

func TestPromoUseCase_BindPromo(t *testing.T) {
	tests := []struct {
		name      string
		mockInit  func(r *repomocks.MockPromoRepository)
		wantError bool
	}{
		{
			name: "успешная привязка",
			mockInit: func(r *repomocks.MockPromoRepository) {
				r.EXPECT().GetPromocodeByCode(gomock.Any(), "SALE").Return(futurePromo(), nil)
				r.EXPECT().BindPromocodeToUser(gomock.Any(), int64(7), int64(1)).Return(true, nil)
			},
		},
		{
			name: "промокод не найден",
			mockInit: func(r *repomocks.MockPromoRepository) {
				r.EXPECT().GetPromocodeByCode(gomock.Any(), "SALE").
					Return(domain.Promocode{}, repository.ErrPromocodeNotFound)
			},
			wantError: true,
		},
		{
			name: "промокод истёк",
			mockInit: func(r *repomocks.MockPromoRepository) {
				p := futurePromo()
				p.ExpiresAt = time.Now().Add(-time.Hour)
				r.EXPECT().GetPromocodeByCode(gomock.Any(), "SALE").Return(p, nil)
			},
			wantError: true,
		},
		{
			name: "уже привязан",
			mockInit: func(r *repomocks.MockPromoRepository) {
				r.EXPECT().GetPromocodeByCode(gomock.Any(), "SALE").Return(futurePromo(), nil)
				r.EXPECT().BindPromocodeToUser(gomock.Any(), int64(7), int64(1)).Return(false, nil)
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := repomocks.NewMockPromoRepository(ctrl)
			tt.mockInit(repo)

			_, err := NewPromoUseCase(repo).BindPromo(context.Background(), 7, "SALE")
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPromoUseCase_ValidatePromo(t *testing.T) {
	tests := []struct {
		name         string
		mockInit     func(r *repomocks.MockPromoRepository)
		wantValid    bool
		wantDiscount int64
	}{
		{
			name: "валиден, фиксированная скидка",
			mockInit: func(r *repomocks.MockPromoRepository) {
				p := futurePromo()
				p.DiscountAmount = ptrInt64(150)
				r.EXPECT().GetPromocodeByCode(gomock.Any(), "SALE").Return(p, nil)
				r.EXPECT().CountUserPromocodeUsage(gomock.Any(), int64(1), int64(7)).Return(0, nil)
			},
			wantValid:    true,
			wantDiscount: 150,
		},
		{
			name: "валиден, процентная скидка",
			mockInit: func(r *repomocks.MockPromoRepository) {
				p := futurePromo()
				p.DiscountPercent = ptrInt(20)
				r.EXPECT().GetPromocodeByCode(gomock.Any(), "SALE").Return(p, nil)
				r.EXPECT().CountUserPromocodeUsage(gomock.Any(), int64(1), int64(7)).Return(0, nil)
			},
			wantValid:    true,
			wantDiscount: 200, // 20% от 1000
		},
		{
			name: "не найден — невалиден без ошибки",
			mockInit: func(r *repomocks.MockPromoRepository) {
				r.EXPECT().GetPromocodeByCode(gomock.Any(), "SALE").
					Return(domain.Promocode{}, repository.ErrPromocodeNotFound)
			},
			wantValid: false,
		},
		{
			name: "сумма заказа ниже минимальной",
			mockInit: func(r *repomocks.MockPromoRepository) {
				p := futurePromo()
				p.DiscountAmount = ptrInt64(150)
				p.MinOrderAmount = ptrInt64(5000)
				r.EXPECT().GetPromocodeByCode(gomock.Any(), "SALE").Return(p, nil)
			},
			wantValid: false,
		},
		{
			name: "уже использован пользователем",
			mockInit: func(r *repomocks.MockPromoRepository) {
				p := futurePromo()
				p.DiscountAmount = ptrInt64(150)
				r.EXPECT().GetPromocodeByCode(gomock.Any(), "SALE").Return(p, nil)
				r.EXPECT().CountUserPromocodeUsage(gomock.Any(), int64(1), int64(7)).Return(1, nil)
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := repomocks.NewMockPromoRepository(ctrl)
			tt.mockInit(repo)

			res, err := NewPromoUseCase(repo).ValidatePromo(context.Background(), 7, "SALE", 0, 1000, 0, 0)
			require.NoError(t, err)
			assert.Equal(t, tt.wantValid, res.Valid)
			if tt.wantValid {
				assert.Equal(t, tt.wantDiscount, res.Discount)
			}
		})
	}
}

func TestPromoUseCase_UsePromo(t *testing.T) {
	t.Run("успешная фиксация", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := repomocks.NewMockPromoRepository(ctrl)
		repo.EXPECT().GetPromocodeByCode(gomock.Any(), "SALE").Return(futurePromo(), nil)
		repo.EXPECT().CountUserPromocodeUsage(gomock.Any(), int64(1), int64(7)).Return(0, nil)
		repo.EXPECT().ResolveOrderInternalID(gomock.Any(), "ord-1").Return(ptrInt64(42), nil)
		repo.EXPECT().RecordPromocodeUsage(gomock.Any(), int64(1), int64(7), ptrInt64(42)).Return(nil)

		err := NewPromoUseCase(repo).UsePromo(context.Background(), 7, "SALE", "ord-1")
		assert.NoError(t, err)
	})

	t.Run("повторный вызов идемпотентен", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := repomocks.NewMockPromoRepository(ctrl)
		repo.EXPECT().GetPromocodeByCode(gomock.Any(), "SALE").Return(futurePromo(), nil)
		repo.EXPECT().CountUserPromocodeUsage(gomock.Any(), int64(1), int64(7)).Return(1, nil)

		err := NewPromoUseCase(repo).UsePromo(context.Background(), 7, "SALE", "ord-1")
		assert.NoError(t, err)
	})

	t.Run("общий лимит исчерпан", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		p := futurePromo()
		p.MaxUses = ptrInt(3)
		p.CurrentUses = 3

		repo := repomocks.NewMockPromoRepository(ctrl)
		repo.EXPECT().GetPromocodeByCode(gomock.Any(), "SALE").Return(p, nil)
		repo.EXPECT().CountUserPromocodeUsage(gomock.Any(), int64(1), int64(7)).Return(0, nil)

		err := NewPromoUseCase(repo).UsePromo(context.Background(), 7, "SALE", "ord-1")
		assert.Error(t, err)
	})

	t.Run("промокод не найден", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := repomocks.NewMockPromoRepository(ctrl)
		repo.EXPECT().GetPromocodeByCode(gomock.Any(), "SALE").
			Return(domain.Promocode{}, repository.ErrPromocodeNotFound)

		err := NewPromoUseCase(repo).UsePromo(context.Background(), 7, "SALE", "ord-1")
		assert.Error(t, err)
	})
}

func TestPromoUseCase_GetUserPromos(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repomocks.NewMockPromoRepository(ctrl)
	repo.EXPECT().GetUserPromocodes(gomock.Any(), int64(7)).
		Return([]domain.Promocode{futurePromo()}, nil)

	promos, err := NewPromoUseCase(repo).GetUserPromos(context.Background(), 7)
	require.NoError(t, err)
	assert.Len(t, promos, 1)

	repo2 := repomocks.NewMockPromoRepository(ctrl)
	repo2.EXPECT().GetUserPromocodes(gomock.Any(), int64(7)).Return(nil, errors.New("db error"))
	_, err = NewPromoUseCase(repo2).GetUserPromos(context.Background(), 7)
	assert.Error(t, err)
}
