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
		idemKey     string
		mockInit    mockInit
		expectedErr error
	}{
		{
			name:    "Успешная блокировка",
			cartID:  "cart-123",
			userID:  1,
			idemKey: "idem-1",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(1), "idem-1", "LockCart").Return(nil)
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
			name:    "Ошибка: юзер не админ корзины",
			cartID:  "cart-123",
			userID:  2,
			idemKey: "idem-2",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(2), "idem-2", "LockCart").Return(nil)
				repo.EXPECT().GetCartByID(gomock.Any(), "cart-123").Return(domain.Cart{
					ID:      "cart-123",
					AdminID: 1,
				}, nil)
			},
			expectedErr: domain.ErrForbidden,
		},
		{
			name:    "Ошибка: есть ничейные (Orphaned) позиции",
			cartID:  "cart-123",
			userID:  1,
			idemKey: "idem-3",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(1), "idem-3", "LockCart").Return(nil)
				repo.EXPECT().GetCartByID(gomock.Any(), "cart-123").Return(domain.Cart{
					ID:      "cart-123",
					AdminID: 1,
					Items:   []domain.CartItem{{DishID: 100, OwnerUserID: nil}},
				}, nil)
			},
			expectedErr: domain.ErrUnassignedItems,
		},
		{
			name:    "Конфликт идемпотентности",
			cartID:  "cart-123",
			userID:  1,
			idemKey: "idem-4",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(1), "idem-4", "LockCart").Return(domain.ErrIdempotencyConflict)
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repoMock := repoMocks.NewMockCartRepository(ctrl)
			tt.mockInit(repoMock)

			uc := NewCartUseCase(repoMock, nil, "")
			err := uc.LockCart(context.Background(), tt.cartID, tt.userID, tt.idemKey)

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
		idemKey     string
		mockInit    mockInit
		expectedErr error
	}{
		{
			name:    "Успешное создание новой корзины и добавление",
			cartID:  "",
			userID:  1,
			dishID:  100,
			qty:     1,
			idemKey: "idem-1",
			mockInit: func(repo *repoMocks.MockCartRepository, restClient *ucMocks.MockRestaurantClient) {
				restClient.EXPECT().GetDishesByIDs(gomock.Any(), []int64{100}).
					Return([]domain.Dish{{ID: 100, RestaurantBrandID: 50}}, nil)

				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(1), "idem-1", "AddItem").Return(nil)

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
			idemKey:     "idem-2",
			mockInit:    func(repo *repoMocks.MockCartRepository, restClient *ucMocks.MockRestaurantClient) {},
			expectedErr: domain.ErrInvalidQuantity,
		},
		{
			name:    "Ошибка: блюдо из другого ресторана",
			cartID:  "cart-1",
			userID:  1,
			dishID:  100,
			qty:     1,
			idemKey: "idem-3",
			mockInit: func(repo *repoMocks.MockCartRepository, restClient *ucMocks.MockRestaurantClient) {
				restClient.EXPECT().GetDishesByIDs(gomock.Any(), []int64{100}).
					Return([]domain.Dish{{ID: 100, RestaurantBrandID: 50}}, nil)

				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(1), "idem-3", "AddItem").Return(nil)

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
		{
			name:    "Конфликт идемпотентности",
			cartID:  "cart-1",
			userID:  1,
			dishID:  100,
			qty:     1,
			idemKey: "idem-4",
			mockInit: func(repo *repoMocks.MockCartRepository, restClient *ucMocks.MockRestaurantClient) {
				restClient.EXPECT().GetDishesByIDs(gomock.Any(), []int64{100}).
					Return([]domain.Dish{{ID: 100, RestaurantBrandID: 50}}, nil)

				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(1), "idem-4", "AddItem").Return(domain.ErrIdempotencyConflict)
			},
			expectedErr: nil,
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
			err := uc.AddItem(context.Background(), tt.cartID, tt.userID, tt.dishID, tt.qty, tt.idemKey)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCartUseCase_RemoveItem(t *testing.T) {
	type mockInit func(repo *repoMocks.MockCartRepository)

	tests := []struct {
		name        string
		cartID      string
		userID      int64
		dishID      int64
		idemKey     string
		mockInit    mockInit
		expectedErr error
	}{
		{
			name:    "Успешное удаление товара",
			cartID:  "cart-1",
			userID:  1,
			dishID:  100,
			idemKey: "idem-remove-1",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				ownerID := int64(1)
				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(1), "idem-remove-1", "RemoveItem").Return(nil)
				repo.EXPECT().GetCartByID(gomock.Any(), "cart-1").Return(domain.Cart{
					ID: "cart-1",
					Items: []domain.CartItem{
						{DishID: 100, OwnerUserID: &ownerID},
					},
				}, nil)
				repo.EXPECT().RemoveItem(gomock.Any(), "cart-1", int64(100), int64(1)).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:    "Товар уже отсутствует (Идемпотентный выход)",
			cartID:  "cart-1",
			userID:  1,
			dishID:  100,
			idemKey: "idem-remove-2",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(1), "idem-remove-2", "RemoveItem").Return(nil)
				repo.EXPECT().GetCartByID(gomock.Any(), "cart-1").Return(domain.Cart{
					ID:    "cart-1",
					Items: []domain.CartItem{}, // Пустая корзина, товара нет
				}, nil)
			},
			expectedErr: nil,
		},
		{
			name:    "Ошибка получения корзины",
			cartID:  "cart-1",
			userID:  1,
			dishID:  100,
			idemKey: "idem-remove-3",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(1), "idem-remove-3", "RemoveItem").Return(nil)
				repo.EXPECT().GetCartByID(gomock.Any(), "cart-1").Return(domain.Cart{}, errors.New("db error"))
			},
			expectedErr: errors.New("db error"),
		},
		{
			name:    "Конфликт идемпотентности",
			cartID:  "cart-1",
			userID:  1,
			dishID:  100,
			idemKey: "idem-remove-4",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(1), "idem-remove-4", "RemoveItem").Return(domain.ErrIdempotencyConflict)
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repoMock := repoMocks.NewMockCartRepository(ctrl)
			tt.mockInit(repoMock)

			uc := NewCartUseCase(repoMock, nil, "")
			err := uc.RemoveItem(context.Background(), tt.cartID, tt.userID, tt.dishID, tt.idemKey)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCartUseCase_UpdateItemQuantity(t *testing.T) {
	type mockInit func(repo *repoMocks.MockCartRepository)

	tests := []struct {
		name        string
		cartID      string
		userID      int64
		dishID      int64
		qty         int32
		idemKey     string
		mockInit    mockInit
		expectedErr error
	}{
		{
			name:    "Успешное обновление",
			cartID:  "cart-1",
			userID:  1,
			dishID:  100,
			qty:     3,
			idemKey: "idem-update-1",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				ownerID := int64(1)
				repo.EXPECT().GetCartByID(gomock.Any(), "cart-1").Return(domain.Cart{
					ID: "cart-1",
					Items: []domain.CartItem{
						{DishID: 100, OwnerUserID: &ownerID},
					},
				}, nil)
				repo.EXPECT().UpdateItemQuantity(gomock.Any(), "cart-1", int64(100), int64(1), int32(3)).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:        "Невалидное количество",
			cartID:      "cart-1",
			userID:      1,
			dishID:      100,
			qty:         0,
			idemKey:     "idem-update-2",
			mockInit:    func(repo *repoMocks.MockCartRepository) {},
			expectedErr: domain.ErrInvalidQuantity,
		},
		{
			name:    "Товар не найден",
			cartID:  "cart-1",
			userID:  1,
			dishID:  100,
			qty:     3,
			idemKey: "idem-update-3",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().GetCartByID(gomock.Any(), "cart-1").Return(domain.Cart{
					ID:    "cart-1",
					Items: []domain.CartItem{},
				}, nil)
			},
			expectedErr: domain.ErrDishNotFound,
		},
		{
			name:    "Ошибка БД",
			cartID:  "cart-1",
			userID:  1,
			dishID:  100,
			qty:     3,
			idemKey: "idem-update-4",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().GetCartByID(gomock.Any(), "cart-1").Return(domain.Cart{}, errors.New("db error"))
			},
			expectedErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repoMock := repoMocks.NewMockCartRepository(ctrl)
			tt.mockInit(repoMock)

			uc := NewCartUseCase(repoMock, nil, "")
			err := uc.UpdateItemQuantity(context.Background(), tt.cartID, tt.userID, tt.dishID, tt.qty, tt.idemKey)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
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

func TestCartUseCase_ReassignItemOwner(t *testing.T) {
	type mockInit func(repo *repoMocks.MockCartRepository)

	tests := []struct {
		name        string
		cartID      string
		adminID     int64
		dishID      int64
		newOwnerID  *int64
		idemKey     string
		mockInit    mockInit
		expectedErr error
	}{
		{
			name:       "Успешное переназначение",
			cartID:     "cart-1",
			adminID:    1,
			dishID:     100,
			newOwnerID: ptr(2),
			idemKey:    "idem-reassign-1",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(1), "idem-reassign-1", "ReassignItemOwner").Return(nil)
				repo.EXPECT().GetCartByID(gomock.Any(), "cart-1").Return(domain.Cart{
					ID:      "cart-1",
					AdminID: 1,
					Members: []domain.CartMember{{UserID: 2}},
				}, nil)
				repo.EXPECT().ReassignItemOwner(gomock.Any(), "cart-1", int64(100), ptr(2)).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:       "Успешное снятие владельца (ничейный товар)",
			cartID:     "cart-1",
			adminID:    1,
			dishID:     100,
			newOwnerID: nil,
			idemKey:    "idem-reassign-2",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(1), "idem-reassign-2", "ReassignItemOwner").Return(nil)
				repo.EXPECT().GetCartByID(gomock.Any(), "cart-1").Return(domain.Cart{
					ID:      "cart-1",
					AdminID: 1,
				}, nil)
				repo.EXPECT().ReassignItemOwner(gomock.Any(), "cart-1", int64(100), nil).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:       "Ошибка: не админ",
			cartID:     "cart-1",
			adminID:    2,
			dishID:     100,
			newOwnerID: ptr(3),
			idemKey:    "idem-reassign-3",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(2), "idem-reassign-3", "ReassignItemOwner").Return(nil)
				repo.EXPECT().GetCartByID(gomock.Any(), "cart-1").Return(domain.Cart{
					ID:      "cart-1",
					AdminID: 1, // Админ другой
				}, nil)
			},
			expectedErr: domain.ErrForbidden,
		},
		{
			name:       "Ошибка: новый владелец не в корзине",
			cartID:     "cart-1",
			adminID:    1,
			dishID:     100,
			newOwnerID: ptr(3),
			idemKey:    "idem-reassign-4",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(1), "idem-reassign-4", "ReassignItemOwner").Return(nil)
				repo.EXPECT().GetCartByID(gomock.Any(), "cart-1").Return(domain.Cart{
					ID:      "cart-1",
					AdminID: 1,
					Members: []domain.CartMember{{UserID: 2}}, // Юзера 3 нет в корзине
				}, nil)
			},
			expectedErr: domain.ErrUserNotInCart,
		},
		{
			name:       "Ошибка получения корзины",
			cartID:     "cart-1",
			adminID:    1,
			dishID:     100,
			newOwnerID: ptr(2),
			idemKey:    "idem-reassign-5",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(1), "idem-reassign-5", "ReassignItemOwner").Return(nil)
				repo.EXPECT().GetCartByID(gomock.Any(), "cart-1").Return(domain.Cart{}, errors.New("db error"))
			},
			expectedErr: errors.New("db error"),
		},
		{
			name:       "Конфликт идемпотентности",
			cartID:     "cart-1",
			adminID:    1,
			dishID:     100,
			newOwnerID: ptr(2),
			idemKey:    "idem-reassign-6",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(1), "idem-reassign-6", "ReassignItemOwner").Return(domain.ErrIdempotencyConflict)
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repoMock := repoMocks.NewMockCartRepository(ctrl)
			tt.mockInit(repoMock)

			uc := NewCartUseCase(repoMock, nil, "")
			err := uc.ReassignItemOwner(context.Background(), tt.cartID, tt.adminID, tt.dishID, tt.newOwnerID, tt.idemKey)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCartUseCase_UnlockCart(t *testing.T) {
	type mockInit func(repo *repoMocks.MockCartRepository)

	tests := []struct {
		name        string
		cartID      string
		userID      int64
		idemKey     string
		mockInit    mockInit
		expectedErr error
	}{
		{
			name:    "Успешная разблокировка",
			cartID:  "cart-1",
			userID:  1,
			idemKey: "idem-unlock-1",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(1), "idem-unlock-1", "UnlockCart").Return(nil)
				repo.EXPECT().GetCartByID(gomock.Any(), "cart-1").Return(domain.Cart{
					ID:      "cart-1",
					Members: []domain.CartMember{{UserID: 1}},
				}, nil)
				repo.EXPECT().UnlockCart(gomock.Any(), "cart-1").Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:    "Ошибка: пользователь не участник",
			cartID:  "cart-1",
			userID:  2,
			idemKey: "idem-unlock-2",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(2), "idem-unlock-2", "UnlockCart").Return(nil)
				repo.EXPECT().GetCartByID(gomock.Any(), "cart-1").Return(domain.Cart{
					ID:      "cart-1",
					Members: []domain.CartMember{{UserID: 1}}, // userID 2 нет
				}, nil)
			},
			expectedErr: domain.ErrForbidden,
		},
		{
			name:    "Конфликт идемпотентности",
			cartID:  "cart-1",
			userID:  1,
			idemKey: "idem-unlock-3",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(1), "idem-unlock-3", "UnlockCart").Return(domain.ErrIdempotencyConflict)
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repoMock := repoMocks.NewMockCartRepository(ctrl)
			tt.mockInit(repoMock)

			uc := NewCartUseCase(repoMock, nil, "")
			err := uc.UnlockCart(context.Background(), tt.cartID, tt.userID, tt.idemKey)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCartUseCase_ClearCart(t *testing.T) {
	type mockInit func(repo *repoMocks.MockCartRepository)

	tests := []struct {
		name        string
		cartID      string
		userID      int64
		idemKey     string
		mockInit    mockInit
		expectedErr error
	}{
		{
			name:    "Успешная очистка",
			cartID:  "cart-1",
			userID:  1,
			idemKey: "idem-clear-1",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(1), "idem-clear-1", "ClearCart").Return(nil)
				repo.EXPECT().GetCartByID(gomock.Any(), "cart-1").Return(domain.Cart{
					ID:      "cart-1",
					AdminID: 1,
				}, nil)
				repo.EXPECT().ClearCart(gomock.Any(), "cart-1").Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:    "Ошибка: пользователь не админ",
			cartID:  "cart-1",
			userID:  2,
			idemKey: "idem-clear-2",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(2), "idem-clear-2", "ClearCart").Return(nil)
				repo.EXPECT().GetCartByID(gomock.Any(), "cart-1").Return(domain.Cart{
					ID:      "cart-1",
					AdminID: 1,
				}, nil)
			},
			expectedErr: domain.ErrForbidden,
		},
		{
			name:    "Конфликт идемпотентности",
			cartID:  "cart-1",
			userID:  1,
			idemKey: "idem-clear-3",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(1), "idem-clear-3", "ClearCart").Return(domain.ErrIdempotencyConflict)
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repoMock := repoMocks.NewMockCartRepository(ctrl)
			tt.mockInit(repoMock)

			uc := NewCartUseCase(repoMock, nil, "")
			err := uc.ClearCart(context.Background(), tt.cartID, tt.userID, tt.idemKey)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCartUseCase_KickMember(t *testing.T) {
	type mockInit func(repo *repoMocks.MockCartRepository)

	tests := []struct {
		name         string
		cartID       string
		adminID      int64
		targetUserID int64
		idemKey      string
		mockInit     mockInit
		expectedErr  error
	}{
		{
			name:         "Успешное исключение",
			cartID:       "cart-1",
			adminID:      1,
			targetUserID: 2,
			idemKey:      "idem-kick-1",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(1), "idem-kick-1", "KickMember").Return(nil)
				repo.EXPECT().GetCartByID(gomock.Any(), "cart-1").Return(domain.Cart{
					ID:      "cart-1",
					AdminID: 1,
				}, nil)
				repo.EXPECT().RemoveMember(gomock.Any(), "cart-1", int64(2)).Return(nil)
				repo.EXPECT().OrphanUserItems(gomock.Any(), "cart-1", int64(2)).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:         "Ошибка: пользователь не админ",
			cartID:       "cart-1",
			adminID:      2,
			targetUserID: 3,
			idemKey:      "idem-kick-2",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(2), "idem-kick-2", "KickMember").Return(nil)
				repo.EXPECT().GetCartByID(gomock.Any(), "cart-1").Return(domain.Cart{
					ID:      "cart-1",
					AdminID: 1,
				}, nil)
			},
			expectedErr: domain.ErrForbidden,
		},
		{
			name:         "Конфликт идемпотентности",
			cartID:       "cart-1",
			adminID:      1,
			targetUserID: 2,
			idemKey:      "idem-kick-3",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(1), "idem-kick-3", "KickMember").Return(domain.ErrIdempotencyConflict)
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repoMock := repoMocks.NewMockCartRepository(ctrl)
			tt.mockInit(repoMock)

			uc := NewCartUseCase(repoMock, nil, "")
			err := uc.KickMember(context.Background(), tt.cartID, tt.adminID, tt.targetUserID, tt.idemKey)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCartUseCase_CloseSharedCart(t *testing.T) {
	type mockInit func(repo *repoMocks.MockCartRepository)

	tests := []struct {
		name        string
		cartID      string
		adminID     int64
		idemKey     string
		mockInit    mockInit
		expectedErr error
	}{
		{
			name:    "Успешное закрытие",
			cartID:  "cart-1",
			adminID: 1,
			idemKey: "idem-close-1",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(1), "idem-close-1", "CloseSharedCart").Return(nil)
				repo.EXPECT().GetCartByID(gomock.Any(), "cart-1").Return(domain.Cart{
					ID:      "cart-1",
					AdminID: 1,
				}, nil)
				repo.EXPECT().DowngradeToSolo(gomock.Any(), "cart-1", int64(1)).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:    "Ошибка: пользователь не админ",
			cartID:  "cart-1",
			adminID: 2,
			idemKey: "idem-close-2",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(2), "idem-close-2", "CloseSharedCart").Return(nil)
				repo.EXPECT().GetCartByID(gomock.Any(), "cart-1").Return(domain.Cart{
					ID:      "cart-1",
					AdminID: 1,
				}, nil)
			},
			expectedErr: domain.ErrForbidden,
		},
		{
			name:    "Конфликт идемпотентности",
			cartID:  "cart-1",
			adminID: 1,
			idemKey: "idem-close-3",
			mockInit: func(repo *repoMocks.MockCartRepository) {
				repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})
				repo.EXPECT().CheckAndSaveIdempotency(gomock.Any(), int64(1), "idem-close-3", "CloseSharedCart").Return(domain.ErrIdempotencyConflict)
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repoMock := repoMocks.NewMockCartRepository(ctrl)
			tt.mockInit(repoMock)

			uc := NewCartUseCase(repoMock, nil, "")
			err := uc.CloseSharedCart(context.Background(), tt.cartID, tt.adminID, tt.idemKey)

			if tt.expectedErr != nil {
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
