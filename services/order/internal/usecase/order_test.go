package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/repository"
	repoMocks "github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/repository/mocks"
	ucMocks "github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/usecase/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
)

type statusCoder interface {
	GRPCStatus() codes.Code
}

type useCaseDeps struct {
	repo *repoMocks.MockOrderRepository
	addr *ucMocks.MockAddressClient
	cart *ucMocks.MockCartClient
	rest *ucMocks.MockRestaurantClient
	pub  *ucMocks.MockMessagePublisher
}

func setupDeps(ctrl *gomock.Controller) useCaseDeps {
	return useCaseDeps{
		repo: repoMocks.NewMockOrderRepository(ctrl),
		addr: ucMocks.NewMockAddressClient(ctrl),
		cart: ucMocks.NewMockCartClient(ctrl),
		rest: ucMocks.NewMockRestaurantClient(ctrl),
		pub:  ucMocks.NewMockMessagePublisher(ctrl),
	}
}

func TestOrderUseCase_CreateOrder(t *testing.T) {
	type mockInit func(d useCaseDeps)

	ownerID := int64(1)
	friendID := int64(2)
	validPromo := "PROMO100"
	idemKey := "idem-1"

	runTx := func(ctx context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	}

	tests := []struct {
		name        string
		req         domain.CreateOrderInput
		mockInit    mockInit
		expectedID  string
		expectedErr bool
		errCode     codes.Code
	}{
		{
			name: "Успешное создание заказа без промокода",
			req: domain.CreateOrderInput{
				UserID:             1,
				AddressPublicID:    "addr-1",
				RestaurantBranchID: 10,
				RestaurantBrandID:  20,
				DeliveryCost:       100,
				ServiceFee:         50,
				PayForAll:          true,
			},
			mockInit: func(d useCaseDeps) {
				d.cart.EXPECT().GetCart(gomock.Any(), int64(1)).Return(domain.Cart{
					ID: "cart-1", Items: []domain.CartItem{{DishID: 1, Quantity: 2, Price: 500}},
				}, int64(1000), nil)

				d.addr.EXPECT().CheckAddressExists(gomock.Any(), int64(1), "addr-1").Return(nil)
				d.rest.EXPECT().GetRestaurantName(gomock.Any(), int64(10)).Return("KFC", nil)
				d.cart.EXPECT().LockCart(gomock.Any(), "cart-1", int64(1), idemKey+"_lock").Return(nil)

				d.repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(runTx)
				d.repo.EXPECT().CreateOrder(gomock.Any(), gomock.Any(), idemKey).Return(int64(1), "pub-uuid", nil)
				d.repo.EXPECT().GetOrderByPublicID(gomock.Any(), "pub-uuid").Return(domain.Order{PublicID: "pub-uuid"}, nil)
				d.pub.EXPECT().PublishJSON(gomock.Any(), events.QueueAnalytics, gomock.Any()).Return(nil)
				d.pub.EXPECT().PublishJSON(gomock.Any(), events.QueuePaymentCommands, gomock.Any()).Return(nil)
			},
			expectedID:  "pub-uuid",
			expectedErr: false,
		},
		{
			name: "Успешное создание со сплитом (PayForAll=false)",
			req: domain.CreateOrderInput{
				UserID:             1,
				AddressPublicID:    "addr-2",
				RestaurantBranchID: 10,
				PayForAll:          false,
				PayerMapping:       map[int64]int64{2: 1},
			},
			mockInit: func(d useCaseDeps) {
				d.cart.EXPECT().GetCart(gomock.Any(), int64(1)).Return(domain.Cart{
					ID: "cart-2",
					Items: []domain.CartItem{
						{DishID: 1, Quantity: 1, Price: 500, OwnerUserID: &ownerID},
						{DishID: 2, Quantity: 1, Price: 300, OwnerUserID: &friendID},
					},
				}, int64(800), nil)

				d.addr.EXPECT().CheckAddressExists(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				d.rest.EXPECT().GetRestaurantName(gomock.Any(), int64(10)).Return("Burger King", nil)
				d.cart.EXPECT().LockCart(gomock.Any(), "cart-2", int64(1), idemKey+"_lock").Return(nil)

				d.repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(runTx)
				d.repo.EXPECT().CreateOrder(gomock.Any(), gomock.Any(), idemKey).Return(int64(2), "pub-uuid-2", nil)
				d.repo.EXPECT().GetOrderByPublicID(gomock.Any(), "pub-uuid-2").Return(domain.Order{PublicID: "pub-uuid-2"}, nil)
				d.pub.EXPECT().PublishJSON(gomock.Any(), events.QueueAnalytics, gomock.Any()).Return(nil)
				d.pub.EXPECT().PublishJSON(gomock.Any(), events.QueuePaymentCommands, gomock.Any()).Return(nil).Times(1)
			},
			expectedID:  "pub-uuid-2",
			expectedErr: false,
		},
		{
			name: "Успешное создание заказа с валидным промокодом (Amount)",
			req: domain.CreateOrderInput{
				UserID:             1,
				AddressPublicID:    "addr-1",
				RestaurantBranchID: 10,
				RestaurantBrandID:  20,
				PayForAll:          true,
				Promocode:          &validPromo,
			},
			mockInit: func(d useCaseDeps) {
				d.cart.EXPECT().GetCart(gomock.Any(), int64(1)).Return(domain.Cart{
					ID: "cart-1", Items: []domain.CartItem{{DishID: 1, Quantity: 2, Price: 500}},
				}, int64(1000), nil)

				d.addr.EXPECT().CheckAddressExists(gomock.Any(), int64(1), "addr-1").Return(nil)
				d.rest.EXPECT().GetRestaurantName(gomock.Any(), int64(10)).Return("KFC", nil)
				d.cart.EXPECT().LockCart(gomock.Any(), "cart-1", int64(1), idemKey+"_lock").Return(nil)

				d.repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(runTx)

				discAmount := int64(200)
				promo := domain.Promocode{
					ID:             1,
					Code:           "PROMO100",
					ExpiresAt:      time.Now().Add(24 * time.Hour),
					DiscountAmount: &discAmount,
				}
				d.repo.EXPECT().GetPromocodeByCodeWithLock(gomock.Any(), "PROMO100").Return(promo, nil)
				d.repo.EXPECT().CheckPromocodeUsage(gomock.Any(), int64(1), int64(1)).Return(false, nil)
				d.repo.EXPECT().CreateOrder(gomock.Any(), gomock.Any(), idemKey).Return(int64(3), "pub-uuid-promo", nil)
				d.repo.EXPECT().IncrementPromocodeUses(gomock.Any(), int64(1)).Return(nil)
				d.repo.EXPECT().CreatePromocodeUsage(gomock.Any(), int64(1), int64(3), int64(1)).Return(nil)
				d.repo.EXPECT().GetOrderByPublicID(gomock.Any(), "pub-uuid-promo").Return(domain.Order{PublicID: "pub-uuid-promo"}, nil)
				d.pub.EXPECT().PublishJSON(gomock.Any(), events.QueueAnalytics, gomock.Any()).Return(nil)
				d.pub.EXPECT().PublishJSON(gomock.Any(), events.QueuePaymentCommands, gomock.Any()).Return(nil)
			},
			expectedID:  "pub-uuid-promo",
			expectedErr: false,
		},
		{
			name: "Ошибка: GetCart вернул ошибку",
			req:  domain.CreateOrderInput{UserID: 1},
			mockInit: func(d useCaseDeps) {
				d.cart.EXPECT().GetCart(gomock.Any(), int64(1)).Return(domain.Cart{}, int64(0), errors.New("db err"))
			},
			expectedErr: true,
			errCode:     codes.Internal,
		},
		{
			name: "Ошибка пустая корзина",
			req:  domain.CreateOrderInput{UserID: 1},
			mockInit: func(d useCaseDeps) {
				d.cart.EXPECT().GetCart(gomock.Any(), int64(1)).Return(domain.Cart{Items: []domain.CartItem{}}, int64(0), nil)
			},
			expectedErr: true,
			errCode:     codes.InvalidArgument,
		},
		{
			name: "Ошибка: неверный адрес",
			req: domain.CreateOrderInput{
				UserID: 1, AddressPublicID: "addr-err",
			},
			mockInit: func(d useCaseDeps) {
				d.cart.EXPECT().GetCart(gomock.Any(), int64(1)).Return(domain.Cart{Items: []domain.CartItem{{DishID: 1}}}, int64(500), nil)
				d.addr.EXPECT().CheckAddressExists(gomock.Any(), int64(1), "addr-err").Return(errors.New("bad addr"))
			},
			expectedErr: true,
			errCode:     codes.InvalidArgument,
		},
		{
			name: "Ошибка: не найден ресторан",
			req: domain.CreateOrderInput{
				UserID: 1, AddressPublicID: "addr-1", RestaurantBranchID: 10,
			},
			mockInit: func(d useCaseDeps) {
				d.cart.EXPECT().GetCart(gomock.Any(), int64(1)).Return(domain.Cart{Items: []domain.CartItem{{DishID: 1}}}, int64(500), nil)
				d.addr.EXPECT().CheckAddressExists(gomock.Any(), int64(1), "addr-1").Return(nil)
				d.rest.EXPECT().GetRestaurantName(gomock.Any(), int64(10)).Return("", errors.New("no rest"))
			},
			expectedErr: true,
			errCode:     codes.Internal,
		},
		{
			name: "Ошибка при блокировке корзины",
			req: domain.CreateOrderInput{
				UserID: 1, AddressPublicID: "addr-1", RestaurantBranchID: 10,
			},
			mockInit: func(d useCaseDeps) {
				d.cart.EXPECT().GetCart(gomock.Any(), int64(1)).Return(domain.Cart{ID: "cart-1", Items: []domain.CartItem{{DishID: 1}}}, int64(500), nil)
				d.addr.EXPECT().CheckAddressExists(gomock.Any(), int64(1), "addr-1").Return(nil)
				d.rest.EXPECT().GetRestaurantName(gomock.Any(), int64(10)).Return("KFC", nil)
				d.cart.EXPECT().LockCart(gomock.Any(), "cart-1", int64(1), idemKey+"_lock").Return(errors.New("lock failed"))
			},
			expectedErr: true,
			errCode:     codes.FailedPrecondition,
		},
		{
			name: "Ошибка: промокод не найден",
			req: domain.CreateOrderInput{
				UserID: 1, AddressPublicID: "addr-1", RestaurantBranchID: 10, Promocode: &validPromo,
			},
			mockInit: func(d useCaseDeps) {
				d.cart.EXPECT().GetCart(gomock.Any(), int64(1)).Return(domain.Cart{ID: "cart-1", Items: []domain.CartItem{{DishID: 1, Quantity: 2}}}, int64(1000), nil)
				d.addr.EXPECT().CheckAddressExists(gomock.Any(), int64(1), "addr-1").Return(nil)
				d.rest.EXPECT().GetRestaurantName(gomock.Any(), int64(10)).Return("KFC", nil)
				d.cart.EXPECT().LockCart(gomock.Any(), "cart-1", int64(1), idemKey+"_lock").Return(nil)

				d.repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(runTx)
				d.repo.EXPECT().GetPromocodeByCodeWithLock(gomock.Any(), "PROMO100").Return(domain.Promocode{}, errors.New("not found"))

				d.cart.EXPECT().UnlockCart(gomock.Any(), "cart-1", int64(1), idemKey+"_unlock").Return(nil)
			},
			expectedErr: true,
			errCode:     codes.NotFound,
		},
		{
			name: "Ошибка: промокод истек",
			req: domain.CreateOrderInput{
				UserID: 1, AddressPublicID: "addr-1", RestaurantBranchID: 10, Promocode: &validPromo,
			},
			mockInit: func(d useCaseDeps) {
				d.cart.EXPECT().GetCart(gomock.Any(), int64(1)).Return(domain.Cart{ID: "cart-1", Items: []domain.CartItem{{DishID: 1}}}, int64(1000), nil)
				d.addr.EXPECT().CheckAddressExists(gomock.Any(), int64(1), "addr-1").Return(nil)
				d.rest.EXPECT().GetRestaurantName(gomock.Any(), int64(10)).Return("KFC", nil)
				d.cart.EXPECT().LockCart(gomock.Any(), "cart-1", int64(1), idemKey+"_lock").Return(nil)

				d.repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(runTx)
				d.repo.EXPECT().GetPromocodeByCodeWithLock(gomock.Any(), "PROMO100").Return(domain.Promocode{
					ExpiresAt: time.Now().Add(-24 * time.Hour),
				}, nil)

				d.cart.EXPECT().UnlockCart(gomock.Any(), "cart-1", int64(1), idemKey+"_unlock").Return(nil)
			},
			expectedErr: true,
			errCode:     codes.FailedPrecondition,
		},
		{
			name: "Ошибка: не хватает минимальной суммы для промокода",
			req: domain.CreateOrderInput{
				UserID: 1, AddressPublicID: "addr-1", RestaurantBranchID: 10, Promocode: &validPromo,
			},
			mockInit: func(d useCaseDeps) {
				d.cart.EXPECT().GetCart(gomock.Any(), int64(1)).Return(domain.Cart{ID: "cart-1", Items: []domain.CartItem{{DishID: 1}}}, int64(1000), nil)
				d.addr.EXPECT().CheckAddressExists(gomock.Any(), int64(1), "addr-1").Return(nil)
				d.rest.EXPECT().GetRestaurantName(gomock.Any(), int64(10)).Return("KFC", nil)
				d.cart.EXPECT().LockCart(gomock.Any(), "cart-1", int64(1), idemKey+"_lock").Return(nil)

				d.repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(runTx)
				minOrder := int64(2000)
				d.repo.EXPECT().GetPromocodeByCodeWithLock(gomock.Any(), "PROMO100").Return(domain.Promocode{
					ExpiresAt: time.Now().Add(24 * time.Hour), MinOrderAmount: &minOrder,
				}, nil)

				d.cart.EXPECT().UnlockCart(gomock.Any(), "cart-1", int64(1), idemKey+"_unlock").Return(nil)
			},
			expectedErr: true,
			errCode:     codes.FailedPrecondition,
		},
		{
			name: "Ошибка: неверный бренд для промокода",
			req: domain.CreateOrderInput{
				UserID: 1, AddressPublicID: "addr-1", RestaurantBranchID: 10, RestaurantBrandID: 20, Promocode: &validPromo,
			},
			mockInit: func(d useCaseDeps) {
				d.cart.EXPECT().GetCart(gomock.Any(), int64(1)).Return(domain.Cart{ID: "cart-1", Items: []domain.CartItem{{DishID: 1}}}, int64(1000), nil)
				d.addr.EXPECT().CheckAddressExists(gomock.Any(), int64(1), "addr-1").Return(nil)
				d.rest.EXPECT().GetRestaurantName(gomock.Any(), int64(10)).Return("KFC", nil)
				d.cart.EXPECT().LockCart(gomock.Any(), "cart-1", int64(1), idemKey+"_lock").Return(nil)

				d.repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(runTx)
				otherBrand := int64(30)
				d.repo.EXPECT().GetPromocodeByCodeWithLock(gomock.Any(), "PROMO100").Return(domain.Promocode{
					ExpiresAt: time.Now().Add(24 * time.Hour), RestaurantBrandID: &otherBrand,
				}, nil)

				d.cart.EXPECT().UnlockCart(gomock.Any(), "cart-1", int64(1), idemKey+"_unlock").Return(nil)
			},
			expectedErr: true,
			errCode:     codes.FailedPrecondition,
		},
		{
			name: "Ошибка: промокод привязан к другому юзеру",
			req: domain.CreateOrderInput{
				UserID: 1, AddressPublicID: "addr-1", RestaurantBranchID: 10, Promocode: &validPromo,
			},
			mockInit: func(d useCaseDeps) {
				d.cart.EXPECT().GetCart(gomock.Any(), int64(1)).Return(domain.Cart{ID: "cart-1", Items: []domain.CartItem{{DishID: 1}}}, int64(1000), nil)
				d.addr.EXPECT().CheckAddressExists(gomock.Any(), int64(1), "addr-1").Return(nil)
				d.rest.EXPECT().GetRestaurantName(gomock.Any(), int64(10)).Return("KFC", nil)
				d.cart.EXPECT().LockCart(gomock.Any(), "cart-1", int64(1), idemKey+"_lock").Return(nil)

				d.repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(runTx)
				otherUser := int64(999)
				d.repo.EXPECT().GetPromocodeByCodeWithLock(gomock.Any(), "PROMO100").Return(domain.Promocode{
					ExpiresAt: time.Now().Add(24 * time.Hour), UserID: &otherUser,
				}, nil)

				d.cart.EXPECT().UnlockCart(gomock.Any(), "cart-1", int64(1), idemKey+"_unlock").Return(nil)
			},
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name: "Ошибка: ошибка проверки использования промокода",
			req: domain.CreateOrderInput{
				UserID: 1, AddressPublicID: "addr-1", RestaurantBranchID: 10, Promocode: &validPromo,
			},
			mockInit: func(d useCaseDeps) {
				d.cart.EXPECT().GetCart(gomock.Any(), int64(1)).Return(domain.Cart{ID: "cart-1", Items: []domain.CartItem{{DishID: 1}}}, int64(1000), nil)
				d.addr.EXPECT().CheckAddressExists(gomock.Any(), int64(1), "addr-1").Return(nil)
				d.rest.EXPECT().GetRestaurantName(gomock.Any(), int64(10)).Return("KFC", nil)
				d.cart.EXPECT().LockCart(gomock.Any(), "cart-1", int64(1), idemKey+"_lock").Return(nil)

				d.repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(runTx)
				d.repo.EXPECT().GetPromocodeByCodeWithLock(gomock.Any(), "PROMO100").Return(domain.Promocode{ExpiresAt: time.Now().Add(24 * time.Hour)}, nil)
				d.repo.EXPECT().CheckPromocodeUsage(gomock.Any(), int64(0), int64(1)).Return(false, errors.New("db fail"))

				d.cart.EXPECT().UnlockCart(gomock.Any(), "cart-1", int64(1), idemKey+"_unlock").Return(nil)
			},
			expectedErr: true,
			errCode:     codes.Internal,
		},
		{
			name: "Ошибка: промокод уже использован",
			req: domain.CreateOrderInput{
				UserID: 1, AddressPublicID: "addr-1", RestaurantBranchID: 10, Promocode: &validPromo,
			},
			mockInit: func(d useCaseDeps) {
				d.cart.EXPECT().GetCart(gomock.Any(), int64(1)).Return(domain.Cart{ID: "cart-1", Items: []domain.CartItem{{DishID: 1}}}, int64(1000), nil)
				d.addr.EXPECT().CheckAddressExists(gomock.Any(), int64(1), "addr-1").Return(nil)
				d.rest.EXPECT().GetRestaurantName(gomock.Any(), int64(10)).Return("KFC", nil)
				d.cart.EXPECT().LockCart(gomock.Any(), "cart-1", int64(1), idemKey+"_lock").Return(nil)

				d.repo.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(runTx)
				d.repo.EXPECT().GetPromocodeByCodeWithLock(gomock.Any(), "PROMO100").Return(domain.Promocode{ExpiresAt: time.Now().Add(24 * time.Hour)}, nil)
				d.repo.EXPECT().CheckPromocodeUsage(gomock.Any(), int64(0), int64(1)).Return(true, nil)

				d.cart.EXPECT().UnlockCart(gomock.Any(), "cart-1", int64(1), idemKey+"_unlock").Return(nil)
			},
			expectedErr: true,
			errCode:     codes.FailedPrecondition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			deps := setupDeps(ctrl)
			tt.mockInit(deps)

			uc := NewOrderUseCase(deps.repo, deps.addr, deps.cart, deps.rest, nil, deps.pub, "http://default-logo", logger.NewNopLogger())

			id, err := uc.CreateOrder(context.Background(), tt.req, idemKey)

			if tt.expectedErr {
				assert.Error(t, err)
				domainErr, ok := err.(statusCoder)
				if ok && tt.errCode != 0 {
					assert.Equal(t, tt.errCode, domainErr.GRPCStatus())
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
}

func TestOrderUseCase_CancelOrder(t *testing.T) {
	type mockInit func(d useCaseDeps)

	tests := []struct {
		name        string
		orderID     string
		userID      int64
		mockInit    mockInit
		expectedErr bool
		errCode     codes.Code
	}{
		{
			name:    "Успешная отмена заказа",
			orderID: "pub-123",
			userID:  1,
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().GetOrderByPublicID(gomock.Any(), "pub-123").Return(domain.Order{
					AdminID: 1, Status: StatusCreated,
				}, nil).Times(2)
				d.repo.EXPECT().UpdateOrderStatus(gomock.Any(), "pub-123", StatusCancelled, StatusCreated, StatusCartLocked, StatusPaymentReady, StatusPaid).Return(nil)
				d.pub.EXPECT().PublishJSON(gomock.Any(), events.QueueAnalytics, gomock.Any()).Return(nil)
				d.repo.EXPECT().RollbackPromocodeUsage(gomock.Any(), "pub-123").Return(nil)
				d.pub.EXPECT().PublishJSON(gomock.Any(), events.QueueGatewayEvents, gomock.Any()).Return(nil)
			},
			expectedErr: false,
		},
		{
			name:    "Ошибка пользователь не владелец",
			orderID: "pub-123",
			userID:  2,
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().GetOrderByPublicID(gomock.Any(), "pub-123").Return(domain.Order{
					AdminID: 1, Status: StatusCreated,
				}, nil)
			},
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name:    "Ошибка: заказ не найден",
			orderID: "pub-123",
			userID:  1,
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().GetOrderByPublicID(gomock.Any(), "pub-123").Return(domain.Order{}, errors.New("db error"))
			},
			expectedErr: true,
			errCode:     codes.NotFound,
		},
		{
			name:    "Ошибка заказ в терминальном статусе",
			orderID: "pub-123",
			userID:  1,
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().GetOrderByPublicID(gomock.Any(), "pub-123").Return(domain.Order{
					AdminID: 1, Status: StatusFinished,
				}, nil).Times(2)
				d.repo.EXPECT().UpdateOrderStatus(gomock.Any(), "pub-123", StatusCancelled, StatusCreated, StatusCartLocked, StatusPaymentReady, StatusPaid).Return(repository.ErrStateChanged)
			},
			expectedErr: true,
			errCode:     codes.FailedPrecondition,
		},
		{
			name:    "Ошибка при обновлении статуса",
			orderID: "pub-123",
			userID:  1,
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().GetOrderByPublicID(gomock.Any(), "pub-123").Return(domain.Order{
					AdminID: 1, Status: StatusCreated,
				}, nil).Times(2)
				d.repo.EXPECT().UpdateOrderStatus(gomock.Any(), "pub-123", StatusCancelled, StatusCreated, StatusCartLocked, StatusPaymentReady, StatusPaid).Return(errors.New("db error"))
			},
			expectedErr: true,
			errCode:     codes.Internal,
		},
		{
			name:    "Ошибка при откате промокода",
			orderID: "pub-123",
			userID:  1,
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().GetOrderByPublicID(gomock.Any(), "pub-123").Return(domain.Order{
					AdminID: 1, Status: StatusCreated,
				}, nil).Times(2)
				d.repo.EXPECT().UpdateOrderStatus(gomock.Any(), "pub-123", StatusCancelled, StatusCreated, StatusCartLocked, StatusPaymentReady, StatusPaid).Return(nil)
				d.pub.EXPECT().PublishJSON(gomock.Any(), events.QueueAnalytics, gomock.Any()).Return(nil)
				d.repo.EXPECT().RollbackPromocodeUsage(gomock.Any(), "pub-123").Return(errors.New("db error"))
			},
			expectedErr: true,
			errCode:     codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			deps := setupDeps(ctrl)
			tt.mockInit(deps)

			uc := NewOrderUseCase(deps.repo, deps.addr, deps.cart, deps.rest, nil, deps.pub, "", logger.NewNopLogger())

			err := uc.CancelOrder(context.Background(), tt.orderID, tt.userID)

			if tt.expectedErr {
				require.Error(t, err)
				domainErr, ok := err.(statusCoder)
				if ok && tt.errCode != 0 {
					assert.Equal(t, tt.errCode, domainErr.GRPCStatus())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOrderUseCase_UpdateOrderStatusByPaymentID(t *testing.T) {
	type mockInit func(d useCaseDeps)

	tests := []struct {
		name        string
		paymentID   string
		status      string
		mockInit    mockInit
		expectedErr bool
	}{
		{
			name:        "Пропуск не-succeeded статуса",
			paymentID:   "pay-1",
			status:      "pending",
			mockInit:    func(d useCaseDeps) {},
			expectedErr: false,
		},
		{
			name:      "Ошибка обновления сплита",
			paymentID: "pay-2",
			status:    "succeeded",
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().UpdateSplitStatusByPaymentID(gomock.Any(), "pay-2", SplitStatusPaid).Return("", "", errors.New("db error"))
			},
			expectedErr: true,
		},
		{
			name:      "Ошибка проверки всех сплитов",
			paymentID: "pay-2",
			status:    "succeeded",
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().UpdateSplitStatusByPaymentID(gomock.Any(), "pay-2", SplitStatusPaid).Return("split-2", "pub-order", nil)
				d.pub.EXPECT().PublishJSON(gomock.Any(), events.QueueGatewayEvents, gomock.Any()).Return(nil)
				d.repo.EXPECT().AreAllSplitsPaid(gomock.Any(), "pub-order").Return(false, errors.New("db error"))
			},
			expectedErr: true,
		},
		{
			name:      "Ошибка обновления заказа (уже обновлен - игнор)",
			paymentID: "pay-2",
			status:    "succeeded",
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().UpdateSplitStatusByPaymentID(gomock.Any(), "pay-2", SplitStatusPaid).Return("split-2", "pub-order", nil)
				d.pub.EXPECT().PublishJSON(gomock.Any(), events.QueueGatewayEvents, gomock.Any()).Return(nil)
				d.repo.EXPECT().AreAllSplitsPaid(gomock.Any(), "pub-order").Return(true, nil)
				d.repo.EXPECT().GetOrderByPublicID(gomock.Any(), "pub-order").Return(domain.Order{PublicID: "pub-order", Status: StatusCartLocked}, nil)
				d.repo.EXPECT().UpdateOrderStatus(gomock.Any(), "pub-order", StatusPaid, StatusCreated, StatusCartLocked, StatusPaymentReady).Return(repository.ErrStateChanged)
			},
			expectedErr: false,
		},
		{
			name:      "Ошибка обновления заказа (БД ошибка)",
			paymentID: "pay-2",
			status:    "succeeded",
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().UpdateSplitStatusByPaymentID(gomock.Any(), "pay-2", SplitStatusPaid).Return("split-2", "pub-order", nil)
				d.pub.EXPECT().PublishJSON(gomock.Any(), events.QueueGatewayEvents, gomock.Any()).Return(nil)
				d.repo.EXPECT().AreAllSplitsPaid(gomock.Any(), "pub-order").Return(true, nil)
				d.repo.EXPECT().GetOrderByPublicID(gomock.Any(), "pub-order").Return(domain.Order{PublicID: "pub-order", Status: StatusCartLocked}, nil)
				d.repo.EXPECT().UpdateOrderStatus(gomock.Any(), "pub-order", StatusPaid, StatusCreated, StatusCartLocked, StatusPaymentReady).Return(errors.New("db error"))
			},
			expectedErr: true,
		},
		{
			name:      "Успех: сплит оплачен, заказ еще нет",
			paymentID: "pay-2",
			status:    "succeeded",
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().UpdateSplitStatusByPaymentID(gomock.Any(), "pay-2", SplitStatusPaid).Return("split-2", "pub-order", nil)
				d.pub.EXPECT().PublishJSON(gomock.Any(), events.QueueGatewayEvents, gomock.Any()).Return(nil)
				d.repo.EXPECT().AreAllSplitsPaid(gomock.Any(), "pub-order").Return(false, nil)
			},
			expectedErr: false,
		},
		{
			name:      "Успех: оплачен последний сплит",
			paymentID: "pay-3",
			status:    "succeeded",
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().UpdateSplitStatusByPaymentID(gomock.Any(), "pay-3", SplitStatusPaid).Return("split-3", "pub-order", nil)
				d.pub.EXPECT().PublishJSON(gomock.Any(), events.QueueGatewayEvents, gomock.Any()).Return(nil)
				d.repo.EXPECT().AreAllSplitsPaid(gomock.Any(), "pub-order").Return(true, nil)
				d.repo.EXPECT().GetOrderByPublicID(gomock.Any(), "pub-order").Return(domain.Order{PublicID: "pub-order", AdminID: 1, RestaurantBranchID: 42, Status: StatusCartLocked}, nil)
				d.repo.EXPECT().UpdateOrderStatus(gomock.Any(), "pub-order", StatusPaid, StatusCreated, StatusCartLocked, StatusPaymentReady).Return(nil)
				d.pub.EXPECT().PublishJSON(gomock.Any(), events.QueueAnalytics, gomock.Any()).Return(nil)
				d.pub.EXPECT().PublishJSON(gomock.Any(), events.QueueGatewayEvents, gomock.Any()).Return(nil)
				d.repo.EXPECT().GetOrderByPublicID(gomock.Any(), "pub-order").Return(domain.Order{PublicID: "pub-order", AdminID: 1, RestaurantBranchID: 42}, nil)
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			deps := setupDeps(ctrl)
			tt.mockInit(deps)

			uc := NewOrderUseCase(deps.repo, deps.addr, deps.cart, deps.rest, nil, deps.pub, "", logger.NewNopLogger())

			err := uc.UpdateOrderStatusByPaymentID(context.Background(), tt.paymentID, tt.status, "idem-key")

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOrderUseCase_ProcessSagaReply(t *testing.T) {
	type mockInit func(d useCaseDeps)

	tests := []struct {
		name        string
		reply       events.SagaReply
		mockInit    mockInit
		expectedErr bool
	}{
		{
			name: "Ошибка саги на шаге PAYMENT -> заказ в payment_ready для ретрая",
			reply: events.SagaReply{
				OrderID: "pub-123", Step: "PAYMENT", Status: events.StatusError, SplitID: "split-1",
			},
			mockInit: func(d useCaseDeps) {
				// Платёж не создался, но заказ валиден: переводим в payment_ready,
				// чтобы фронт мог предложить повторную оплату/другую карту.
				d.repo.EXPECT().UpdateOrderStatus(gomock.Any(), "pub-123", StatusFailed, StatusCreated, StatusCartLocked, StatusPaymentReady).Return(nil)
				d.repo.EXPECT().UpdateSplitStatus(gomock.Any(), "split-1", SplitStatusFailed).Return(nil)
				d.repo.EXPECT().RollbackPromocodeUsage(gomock.Any(), "pub-123").Return(nil)
				d.repo.EXPECT().UpdateOrderStatus(gomock.Any(), "pub-123", StatusPaymentReady).Return(nil)
				d.repo.EXPECT().UpdateSplitStatus(gomock.Any(), "split-1", SplitStatusFailed).Return(nil)
				// Событие для фронта через gateway-очередь.
				d.pub.EXPECT().PublishJSON(gomock.Any(), events.QueueGatewayEvents, gomock.Any()).Return(nil)
			},
			expectedErr: false,
		},
		{
			name: "Ошибка саги -> ошибка отката промокода",
			reply: events.SagaReply{
				OrderID: "pub-123", Step: "PAYMENT", Status: events.StatusError, SplitID: "split-1",
			},
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().UpdateOrderStatus(gomock.Any(), "pub-123", StatusFailed, StatusCreated, StatusCartLocked, StatusPaymentReady).Return(nil)
				d.repo.EXPECT().UpdateSplitStatus(gomock.Any(), "split-1", SplitStatusFailed).Return(nil)
				d.repo.EXPECT().RollbackPromocodeUsage(gomock.Any(), "pub-123").Return(errors.New("db error"))
			},
			expectedErr: true,
		},
		{
			name: "Успешный шаг PAYMENT -> сохранение ID и отправка в Gateway",
			reply: events.SagaReply{
				OrderID: "pub-123", Step: "PAYMENT", Status: events.StatusSuccess, SplitID: "split-1", PaymentID: "yoo-123", PaymentURL: "http://pay",
			},
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().SetSplitYookassaID(gomock.Any(), "split-1", "yoo-123").Return(nil)
				d.pub.EXPECT().PublishJSON(gomock.Any(), events.QueueGatewayEvents, gomock.Any()).Return(nil)
			},
			expectedErr: false,
		},
		{
			name: "Успех, но неизвестный шаг (игнорируется)",
			reply: events.SagaReply{
				OrderID: "pub-123", Step: "UNKNOWN", Status: events.StatusSuccess,
			},
			mockInit:    func(d useCaseDeps) {},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			deps := setupDeps(ctrl)
			tt.mockInit(deps)

			uc := NewOrderUseCase(deps.repo, deps.addr, deps.cart, deps.rest, nil, deps.pub, "", logger.NewNopLogger())

			err := uc.ProcessSagaReply(context.Background(), tt.reply)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOrderUseCase_GetOrders(t *testing.T) {
	type mockInit func(d useCaseDeps)

	tests := []struct {
		name        string
		userID      int64
		limit       int32
		offset      int32
		mockInit    mockInit
		expectedRes []domain.Order
		expectedErr bool
	}{
		{
			name:   "Успешное получение заказов с пагинацией",
			userID: 1,
			limit:  10,
			offset: 0,
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().GetOrdersByUserID(gomock.Any(), int64(1), int32(10), int32(0)).Return([]domain.Order{
					{ID: 1, RestaurantBrandID: 10},
					{ID: 2, RestaurantBrandID: 20},
				}, nil)
			},
			expectedRes: []domain.Order{
				{ID: 1, RestaurantBrandID: 10},
				{ID: 2, RestaurantBrandID: 20},
			},
			expectedErr: false,
		},
		{
			name:   "Пустой список заказов",
			userID: 1,
			limit:  10,
			offset: 10,
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().GetOrdersByUserID(gomock.Any(), int64(1), int32(10), int32(10)).Return([]domain.Order{}, nil)
			},
			expectedRes: []domain.Order{},
			expectedErr: false,
		},
		{
			name:   "Ошибка репозитория",
			userID: 1,
			limit:  5,
			offset: 0,
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().GetOrdersByUserID(gomock.Any(), int64(1), int32(5), int32(0)).Return(nil, errors.New("db error"))
			},
			expectedRes: []domain.Order{},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			deps := setupDeps(ctrl)
			tt.mockInit(deps)

			uc := NewOrderUseCase(deps.repo, deps.addr, deps.cart, deps.rest, nil, deps.pub, "http://default-logo", logger.NewNopLogger())

			orders, err := uc.GetOrders(context.Background(), tt.userID, tt.limit, tt.offset)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, orders)
			}
		})
	}
}

func TestOrderUseCase_PayForFriend(t *testing.T) {
	type mockInit func(d useCaseDeps)

	tests := []struct {
		name            string
		splitID         string
		adminID         int64
		paymentMethodID string
		idemKey         string
		mockInit        mockInit
		expectedErr     bool
		errCode         codes.Code
	}{
		{
			name:            "Успешная оплата за друга",
			splitID:         "split-1",
			adminID:         1,
			paymentMethodID: "pm-1",
			idemKey:         "idem-1",
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().GetSplitByID(gomock.Any(), "split-1").Return(domain.OrderSplit{
					ID: "split-1", OrderID: 100, Status: SplitStatusPending, Amount: 500,
				}, nil)
				d.repo.EXPECT().UpdateSplitPayer(gomock.Any(), "split-1", int64(1)).Return(nil)

				d.pub.EXPECT().PublishJSON(gomock.Any(), events.QueuePaymentCommands, gomock.Any()).Return(nil)
			},
			expectedErr: false,
		},
		{
			name:            "Ошибка: сплит не найден",
			splitID:         "split-unknown",
			adminID:         1,
			paymentMethodID: "pm-1",
			idemKey:         "idem-1",
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().GetSplitByID(gomock.Any(), "split-unknown").Return(domain.OrderSplit{}, errors.New("not found"))
			},
			expectedErr: true,
			errCode:     codes.NotFound,
		},
		{
			name:            "Ошибка: сплит не в статусе pending",
			splitID:         "split-2",
			adminID:         1,
			paymentMethodID: "pm-1",
			idemKey:         "idem-1",
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().GetSplitByID(gomock.Any(), "split-2").Return(domain.OrderSplit{
					ID: "split-2", OrderID: 100, Status: SplitStatusPaid, Amount: 500,
				}, nil)
			},
			expectedErr: true,
			errCode:     codes.FailedPrecondition,
		},
		{
			name:            "Ошибка: не удалось обновить плательщика",
			splitID:         "split-3",
			adminID:         2,
			paymentMethodID: "pm-1",
			idemKey:         "idem-1",
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().GetSplitByID(gomock.Any(), "split-3").Return(domain.OrderSplit{
					ID: "split-3", OrderID: 100, Status: SplitStatusPending, Amount: 500,
				}, nil)
				d.repo.EXPECT().UpdateSplitPayer(gomock.Any(), "split-3", int64(2)).Return(errors.New("db error"))
			},
			expectedErr: true,
			errCode:     codes.InvalidArgument,
		},
		{
			name:            "Ошибка при публикации в RabbitMQ",
			splitID:         "split-4",
			adminID:         1,
			paymentMethodID: "pm-1",
			idemKey:         "idem-1",
			mockInit: func(d useCaseDeps) {
				d.repo.EXPECT().GetSplitByID(gomock.Any(), "split-4").Return(domain.OrderSplit{
					ID: "split-4", OrderID: 100, Status: SplitStatusPending, Amount: 500,
				}, nil)
				d.repo.EXPECT().UpdateSplitPayer(gomock.Any(), "split-4", int64(1)).Return(nil)
				d.pub.EXPECT().PublishJSON(gomock.Any(), events.QueuePaymentCommands, gomock.Any()).Return(errors.New("rabbit error"))
			},
			expectedErr: true,
			errCode:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			deps := setupDeps(ctrl)
			tt.mockInit(deps)

			uc := NewOrderUseCase(deps.repo, deps.addr, deps.cart, deps.rest, nil, deps.pub, "http://default-logo", logger.NewNopLogger())

			err := uc.PayForFriend(context.Background(), tt.splitID, tt.adminID, tt.paymentMethodID, tt.idemKey)

			if tt.expectedErr {
				require.Error(t, err)
				domainErr, ok := err.(statusCoder)
				if ok && tt.errCode != 0 {
					assert.Equal(t, tt.errCode, domainErr.GRPCStatus())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
