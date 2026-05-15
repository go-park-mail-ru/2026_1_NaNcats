package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/domain"
	repoMocks "github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/repository/mocks"
	ucMocks "github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func ptr(i int64) *int64 {
	return &i
}

func TestCartUseCase_GetCart(t *testing.T) {
	type mockInit func(repo *repoMocks.MockCartRepository, restClient *ucMocks.MockRestaurantClient)

	tests := []struct {
		name             string
		userID           int64
		mockInit         mockInit
		expectedItemsLen int
		expectedCost     int64
		expectedErr      error
	}{
		{
			name:   "Успешное получение пустой корзины",
			userID: 1,
			mockInit: func(repo *repoMocks.MockCartRepository, restClient *ucMocks.MockRestaurantClient) {
				repo.EXPECT().GetCartByUserID(gomock.Any(), int64(1)).
					Return(domain.Cart{ID: "cart-1", Items: nil}, nil)
			},
			expectedItemsLen: 0,
			expectedCost:     0,
			expectedErr:      nil,
		},
		{
			name:   "Успешное получение обогащенной корзины",
			userID: 1,
			mockInit: func(repo *repoMocks.MockCartRepository, restClient *ucMocks.MockRestaurantClient) {
				dbCart := domain.Cart{
					ID: "cart-1",
					Items: []domain.CartItem{
						{DishID: 100, Quantity: 2},
						{DishID: 200, Quantity: 1},
					},
				}
				repo.EXPECT().GetCartByUserID(gomock.Any(), int64(1)).Return(dbCart, nil)

				restClient.EXPECT().GetDishesByIDs(gomock.Any(), []int64{100, 200}).
					Return([]domain.Dish{
						{ID: 100, Name: "Burger", Price: 500, ImageURL: "url1"},
						{ID: 200, Name: "Cola", Price: 100, ImageURL: "url2"},
					}, nil)
			},
			expectedItemsLen: 2,
			expectedCost:     1100,
			expectedErr:      nil,
		},
		{
			name:   "Деградация (Fallback): Restaurant Client упал",
			userID: 1,
			mockInit: func(repo *repoMocks.MockCartRepository, restClient *ucMocks.MockRestaurantClient) {
				dbCart := domain.Cart{
					ID: "cart-1",
					Items: []domain.CartItem{
						{DishID: 100, Quantity: 2},
					},
				}
				repo.EXPECT().GetCartByUserID(gomock.Any(), int64(1)).Return(dbCart, nil)

				restClient.EXPECT().GetDishesByIDs(gomock.Any(), []int64{100}).
					Return(nil, errors.New("restaurant service timeout"))
			},
			expectedItemsLen: 1,
			expectedCost:     0,
			expectedErr:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repoMock := repoMocks.NewMockCartRepository(ctrl)
			restClientMock := ucMocks.NewMockRestaurantClient(ctrl)
			tt.mockInit(repoMock, restClientMock)

			uc := NewCartUseCase(repoMock, restClientMock, "default.png")
			cart, cost, err := uc.GetCart(context.Background(), tt.userID)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Len(t, cart.Items, tt.expectedItemsLen)
				assert.Equal(t, tt.expectedCost, cost)
			}
		})
	}
}

func TestCartUseCase_LockCart(t *testing.T) {
	type mockInit func(repo *repoMocks.MockCartRepository)

	tests := []struct {
		name        string
		cartID      string
		userID      int64
		mockInit    mockInit
		expectedErr error
	}{
		{
			name:   "Успешная блокировка и очистка корзины",
			cartID: "cart-123",
			userID: 1,
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().GetCartByID(gomock.Any(), "cart-123").Return(domain.Cart{
					ID:      "cart-123",
					AdminID: 1,
					Items:   []domain.CartItem{{DishID: 100, OwnerUserID: ptr(1)}},
				}, nil)

				repo.EXPECT().LockCart(gomock.Any(), "cart-123").Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:   "Ошибка: юзер не админ корзины",
			cartID: "cart-123",
			userID: 2,
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().GetCartByID(gomock.Any(), "cart-123").Return(domain.Cart{
					ID:      "cart-123",
					AdminID: 1,
				}, nil)
			},
			expectedErr: domain.ErrForbidden,
		},
		{
			name:   "Ошибка: есть ничейные (Orphaned) позиции",
			cartID: "cart-123",
			userID: 1,
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().GetCartByID(gomock.Any(), "cart-123").Return(domain.Cart{
					ID:      "cart-123",
					AdminID: 1,
					Items:   []domain.CartItem{{DishID: 100, OwnerUserID: nil}},
				}, nil)
			},
			expectedErr: domain.ErrUnassignedItems,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repoMock := repoMocks.NewMockCartRepository(ctrl)
			tt.mockInit(repoMock)

			uc := NewCartUseCase(repoMock, nil, "")
			err := uc.LockCart(context.Background(), tt.cartID, tt.userID, "idem-1")

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCartUseCase_AddItem(t *testing.T) {
	type mockInit func(repo *repoMocks.MockCartRepository, restClient *ucMocks.MockRestaurantClient)

	tests := []struct {
		name        string
		cartID      string
		userID      int64
		dishID      int64
		qty         int32
		mockInit    mockInit
		expectedErr error
	}{
		{
			name:   "Успешное создание новой корзины и добавление",
			cartID: "",
			userID: 1,
			dishID: 100,
			qty:    1,
			mockInit: func(repo *repoMocks.MockCartRepository, restClient *ucMocks.MockRestaurantClient) {
				restClient.EXPECT().GetDishesByIDs(gomock.Any(), []int64{100}).
					Return([]domain.Dish{{ID: 100, RestaurantBrandID: 50}}, nil)

				repo.EXPECT().GetActiveCartByUserID(gomock.Any(), int64(1)).
					Return(domain.Cart{}, errors.New("not found"))

				repo.EXPECT().CreateCart(gomock.Any(), int64(1), int64(50)).
					Return("new-cart-1", nil)

				expectedItem := domain.CartItem{DishID: 100, Quantity: 1, OwnerUserID: ptr(1)}
				repo.EXPECT().AddItem(gomock.Any(), "new-cart-1", expectedItem).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:        "Ошибка: невалидное количество",
			cartID:      "cart-1",
			userID:      1,
			dishID:      100,
			qty:         0,
			mockInit:    func(repo *repoMocks.MockCartRepository, restClient *ucMocks.MockRestaurantClient) {},
			expectedErr: domain.ErrInvalidQuantity,
		},
		{
			name:   "Ошибка: блюдо из другого ресторана",
			cartID: "cart-1",
			userID: 1,
			dishID: 100,
			qty:    1,
			mockInit: func(repo *repoMocks.MockCartRepository, restClient *ucMocks.MockRestaurantClient) {
				restClient.EXPECT().GetDishesByIDs(gomock.Any(), []int64{100}).
					Return([]domain.Dish{{ID: 100, RestaurantBrandID: 50}}, nil)

				repo.EXPECT().GetCartByID(gomock.Any(), "cart-1").Return(domain.Cart{
					ID:                "cart-1",
					RestaurantBrandID: 20,
					Status:            domain.CartStatusActive,
					Members:           []domain.CartMember{{UserID: 1}},
					Items:             []domain.CartItem{{DishID: 200}},
				}, nil)
			},
			expectedErr: domain.ErrMultipleRestaurants,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repoMock := repoMocks.NewMockCartRepository(ctrl)
			restClientMock := ucMocks.NewMockRestaurantClient(ctrl)
			tt.mockInit(repoMock, restClientMock)

			uc := NewCartUseCase(repoMock, restClientMock, "")
			err := uc.AddItem(context.Background(), tt.cartID, tt.userID, tt.dishID, tt.qty, "idem-1")

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCartUseCase_RoomManagement(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := repoMocks.NewMockCartRepository(ctrl)
	uc := NewCartUseCase(repoMock, nil, "")
	ctx := context.Background()

	t.Run("GenerateInvite - Перевод в Shared и создание", func(t *testing.T) {
		repoMock.EXPECT().GetCartByID(gomock.Any(), "cart-1").Return(domain.Cart{
			ID: "cart-1", AdminID: 1, Mode: domain.CartModeSolo,
		}, nil)

		repoMock.EXPECT().UpdateCartMode(gomock.Any(), "cart-1", domain.CartModeShared).Return(nil)
		repoMock.EXPECT().SaveInvite(gomock.Any(), gomock.Any()).Return(nil)

		invite, err := uc.GenerateInvite(ctx, "cart-1", 1)
		assert.NoError(t, err)
		assert.NotEmpty(t, invite.Token)
	})

	t.Run("JoinCart - Успешный вход", func(t *testing.T) {
		repoMock.EXPECT().GetInviteByToken(gomock.Any(), "token-123").Return(domain.CartInvite{
			Token: "token-123", CartID: "cart-1", ExpiresAt: time.Now().Add(1 * time.Hour),
		}, nil)

		repoMock.EXPECT().AddMember(gomock.Any(), "cart-1", int64(2)).Return(nil)

		cartID, err := uc.JoinCart(ctx, "token-123", 2)
		assert.NoError(t, err)
		assert.Equal(t, "cart-1", cartID)
	})

	t.Run("JoinCart - Токен просрочен", func(t *testing.T) {
		repoMock.EXPECT().GetInviteByToken(gomock.Any(), "token-exp").Return(domain.CartInvite{
			Token: "token-exp", CartID: "cart-1", ExpiresAt: time.Now().Add(-1 * time.Hour),
		}, nil)

		_, err := uc.JoinCart(ctx, "token-exp", 2)
		assert.ErrorIs(t, err, domain.ErrInviteExpired)
	})
}
