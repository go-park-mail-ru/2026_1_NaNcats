package cart

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/delivery/websocket"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/cartclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/cartclient/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/userclient"
	userMocks "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/userclient/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// makeHandlerForKickTest строит хэндлер, где userClient.ResolvePublicID возвращает
// фиксированный internalID=42 при resolveErr==nil или указанную ошибку.
func makeHandlerForKickTest(t *testing.T, ctrl *gomock.Controller, cartMock *mocks.MockCartClient, resolveErr error) *CartHandler {
	t.Helper()
	userMock := userMocks.NewMockUserClient(ctrl)
	userMock.EXPECT().GetUsersByIDs(gomock.Any(), gomock.Any()).Return(nil, errors.New("mock fallback")).AnyTimes()
	if resolveErr != nil {
		userMock.EXPECT().ResolvePublicID(gomock.Any(), gomock.Any()).Return(int64(0), resolveErr).AnyTimes()
	} else {
		userMock.EXPECT().ResolvePublicID(gomock.Any(), gomock.Any()).Return(int64(42), nil).AnyTimes()
	}
	return NewCartHandler(cartMock, userMock, (*websocket.WsManager)(nil), logger.NewNopLogger())
}

// doJSON собирает HTTP-запрос с заданным телом, заголовками и опционально подставляет userID в контекст.
func doJSON(method, path string, body any, headers map[string]string, withAuth bool, userID int64, h func(http.ResponseWriter, *http.Request)) *httptest.ResponseRecorder {
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewBuffer(raw))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if withAuth {
		req = withUserIDContext(req, userID)
	}
	w := httptest.NewRecorder()
	h(w, req)
	return w
}

func TestCartHandler_AddItem_AllErrorBranches(t *testing.T) {
	type mockBehavior func(m *mocks.MockCartClient)

	validBody := AddItemRequest{CartID: "cart-1", DishID: 5, Quantity: 2}

	tests := []struct {
		name           string
		body           any
		headers        map[string]string
		withAuth       bool
		mockBehavior   mockBehavior
		expectedStatus int
	}{
		{
			name:           "Без Idempotency-Key — 400",
			body:           validBody,
			headers:        map[string]string{},
			withAuth:       true,
			mockBehavior:   func(m *mocks.MockCartClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Битый JSON — 400",
			body:           "not a json struct",
			headers:        map[string]string{"Idempotency-Key": "k1"},
			withAuth:       true,
			mockBehavior:   func(m *mocks.MockCartClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "MULTIPLE_RESTAURANTS — 409",
			body:     validBody,
			headers:  map[string]string{"Idempotency-Key": "k2"},
			withAuth: true,
			mockBehavior: func(m *mocks.MockCartClient) {
				m.EXPECT().AddItem(gomock.Any(), "cart-1", int64(1), int64(5), int32(2), "k2").
					Return(cartclient.ErrMultipleRestaurants)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:     "CART_LOCKED — 409",
			body:     validBody,
			headers:  map[string]string{"Idempotency-Key": "k3"},
			withAuth: true,
			mockBehavior: func(m *mocks.MockCartClient) {
				m.EXPECT().AddItem(gomock.Any(), "cart-1", int64(1), int64(5), int32(2), "k3").
					Return(cartclient.ErrCartLocked)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:     "Forbidden — 403",
			body:     validBody,
			headers:  map[string]string{"Idempotency-Key": "k4"},
			withAuth: true,
			mockBehavior: func(m *mocks.MockCartClient) {
				m.EXPECT().AddItem(gomock.Any(), "cart-1", int64(1), int64(5), int32(2), "k4").
					Return(cartclient.ErrForbidden)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:     "InvalidQuantity — 400",
			body:     validBody,
			headers:  map[string]string{"Idempotency-Key": "k5"},
			withAuth: true,
			mockBehavior: func(m *mocks.MockCartClient) {
				m.EXPECT().AddItem(gomock.Any(), "cart-1", int64(1), int64(5), int32(2), "k5").
					Return(cartclient.ErrInvalidQuantity)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "InvalidCart — 400",
			body:     validBody,
			headers:  map[string]string{"Idempotency-Key": "k6"},
			withAuth: true,
			mockBehavior: func(m *mocks.MockCartClient) {
				m.EXPECT().AddItem(gomock.Any(), "cart-1", int64(1), int64(5), int32(2), "k6").
					Return(cartclient.ErrInvalidCart)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Внутренняя ошибка — 500",
			body:     validBody,
			headers:  map[string]string{"Idempotency-Key": "k7"},
			withAuth: true,
			mockBehavior: func(m *mocks.MockCartClient) {
				m.EXPECT().AddItem(gomock.Any(), "cart-1", int64(1), int64(5), int32(2), "k7").
					Return(errors.New("boom"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			h, mc := setupTestHandler(ctrl)
			tt.mockBehavior(mc)
			w := doJSON(http.MethodPost, "/api/cart/items", tt.body, tt.headers, tt.withAuth, 1, h.AddItem)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCartHandler_GenerateInvite(t *testing.T) {
	tests := []struct {
		name           string
		withAuth       bool
		setMock        func(m *mocks.MockCartClient)
		expectedStatus int
	}{
		{
			name:           "Неавторизован — 401",
			withAuth:       false,
			setMock:        func(m *mocks.MockCartClient) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:     "Успех",
			withAuth: true,
			setMock: func(m *mocks.MockCartClient) {
				m.EXPECT().GenerateInvite(gomock.Any(), "cart-1", int64(1)).
					Return(&cartclient.InviteResponse{Token: "tok", ExpiresAt: "2026-12-31"}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "Forbidden — 403",
			withAuth: true,
			setMock: func(m *mocks.MockCartClient) {
				m.EXPECT().GenerateInvite(gomock.Any(), "cart-1", int64(1)).Return(nil, cartclient.ErrForbidden)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:     "Прочая ошибка — 500",
			withAuth: true,
			setMock: func(m *mocks.MockCartClient) {
				m.EXPECT().GenerateInvite(gomock.Any(), "cart-1", int64(1)).Return(nil, errors.New("x"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			h, mc := setupTestHandler(ctrl)
			tt.setMock(mc)

			req := httptest.NewRequest(http.MethodPost, "/api/cart/invite?cart_id=cart-1", nil)
			if tt.withAuth {
				req = withUserIDContext(req, 1)
			}
			w := httptest.NewRecorder()
			h.GenerateInvite(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCartHandler_RemoveItem(t *testing.T) {
	validBody := RemoveItemRequest{CartID: "c", DishID: 5}
	tests := []struct {
		name           string
		body           any
		headers        map[string]string
		withAuth       bool
		setMock        func(m *mocks.MockCartClient)
		expectedStatus int
	}{
		{
			name:           "Неавторизован — 401",
			body:           validBody,
			headers:        map[string]string{"Idempotency-Key": "k"},
			withAuth:       false,
			setMock:        func(m *mocks.MockCartClient) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Без Idempotency-Key — 400",
			body:           validBody,
			headers:        map[string]string{},
			withAuth:       true,
			setMock:        func(m *mocks.MockCartClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Битый JSON — 400",
			body:           "bad",
			headers:        map[string]string{"Idempotency-Key": "k"},
			withAuth:       true,
			setMock:        func(m *mocks.MockCartClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Успех",
			body:     validBody,
			headers:  map[string]string{"Idempotency-Key": "k1"},
			withAuth: true,
			setMock: func(m *mocks.MockCartClient) {
				m.EXPECT().RemoveItem(gomock.Any(), "c", int64(1), int64(5), "k1").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "Forbidden — 403",
			body:     validBody,
			headers:  map[string]string{"Idempotency-Key": "k2"},
			withAuth: true,
			setMock: func(m *mocks.MockCartClient) {
				m.EXPECT().RemoveItem(gomock.Any(), "c", int64(1), int64(5), "k2").Return(cartclient.ErrForbidden)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:     "Internal — 500",
			body:     validBody,
			headers:  map[string]string{"Idempotency-Key": "k3"},
			withAuth: true,
			setMock: func(m *mocks.MockCartClient) {
				m.EXPECT().RemoveItem(gomock.Any(), "c", int64(1), int64(5), "k3").Return(errors.New("x"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			h, mc := setupTestHandler(ctrl)
			tt.setMock(mc)
			w := doJSON(http.MethodDelete, "/api/cart/items", tt.body, tt.headers, tt.withAuth, 1, h.RemoveItem)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCartHandler_UpdateQuantity(t *testing.T) {
	validBody := UpdateQuantityRequest{CartID: "c", DishID: 5, Quantity: 3}
	tests := []struct {
		name           string
		body           any
		headers        map[string]string
		withAuth       bool
		setMock        func(m *mocks.MockCartClient)
		expectedStatus int
	}{
		{
			name: "Неавторизован — 401", body: validBody, headers: map[string]string{"Idempotency-Key": "k"}, withAuth: false,
			setMock:        func(m *mocks.MockCartClient) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Без Idempotency-Key — 400", body: validBody, headers: map[string]string{}, withAuth: true,
			setMock:        func(m *mocks.MockCartClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Битый JSON — 400", body: "x", headers: map[string]string{"Idempotency-Key": "k"}, withAuth: true,
			setMock:        func(m *mocks.MockCartClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Успех", body: validBody, headers: map[string]string{"Idempotency-Key": "k1"}, withAuth: true,
			setMock: func(m *mocks.MockCartClient) {
				m.EXPECT().UpdateItemQuantity(gomock.Any(), "c", int64(1), int64(5), int32(3), "k1").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "InvalidCart — 400", body: validBody, headers: map[string]string{"Idempotency-Key": "k2"}, withAuth: true,
			setMock: func(m *mocks.MockCartClient) {
				m.EXPECT().UpdateItemQuantity(gomock.Any(), "c", int64(1), int64(5), int32(3), "k2").Return(cartclient.ErrInvalidCart)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Forbidden — 403", body: validBody, headers: map[string]string{"Idempotency-Key": "k3"}, withAuth: true,
			setMock: func(m *mocks.MockCartClient) {
				m.EXPECT().UpdateItemQuantity(gomock.Any(), "c", int64(1), int64(5), int32(3), "k3").Return(cartclient.ErrForbidden)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "Internal — 500", body: validBody, headers: map[string]string{"Idempotency-Key": "k4"}, withAuth: true,
			setMock: func(m *mocks.MockCartClient) {
				m.EXPECT().UpdateItemQuantity(gomock.Any(), "c", int64(1), int64(5), int32(3), "k4").Return(errors.New("x"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			h, mc := setupTestHandler(ctrl)
			tt.setMock(mc)
			w := doJSON(http.MethodPut, "/api/cart/items", tt.body, tt.headers, tt.withAuth, 1, h.UpdateQuantity)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCartHandler_ReassignOwner(t *testing.T) {
	publicID := "uuid-target"
	bodyWithOwner := ReassignOwnerRequest{CartID: "c", DishID: 5, NewOwnerPublicID: &publicID}
	bodyNoOwner := ReassignOwnerRequest{CartID: "c", DishID: 5, NewOwnerPublicID: nil}

	tests := []struct {
		name           string
		body           any
		headers        map[string]string
		withAuth       bool
		setMock        func(c *mocks.MockCartClient, u *mocks.MockCartClient)
		expectedStatus int
	}{
		{
			name:     "Без Idempotency-Key — 400",
			body:     bodyWithOwner,
			headers:  map[string]string{},
			withAuth: true,
			setMock:  func(c, u *mocks.MockCartClient) {},
			// userClient.ResolvePublicID не должен быть вызван
			expectedStatus: http.StatusBadRequest,
		},
	}
	_ = tests

	// Полные сценарии прогоняем без таблицы — нужен отдельный мок userClient.
	t.Run("Неавторизован — 401", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h, _ := setupTestHandler(ctrl)
		w := doJSON(http.MethodPatch, "/api/cart/items/owner", bodyWithOwner, map[string]string{"Idempotency-Key": "k"}, false, 1, h.ReassignOwner)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Без Idempotency-Key — 400", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h, _ := setupTestHandler(ctrl)
		w := doJSON(http.MethodPatch, "/api/cart/items/owner", bodyWithOwner, nil, true, 1, h.ReassignOwner)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Битый JSON — 400", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h, _ := setupTestHandler(ctrl)
		w := doJSON(http.MethodPatch, "/api/cart/items/owner", "bad", map[string]string{"Idempotency-Key": "k"}, true, 1, h.ReassignOwner)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Без нового владельца — успех", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h, mc := setupTestHandler(ctrl)
		mc.EXPECT().ReassignItemOwner(gomock.Any(), "c", int64(1), int64(5), gomock.Nil(), "k").Return(nil)
		w := doJSON(http.MethodPatch, "/api/cart/items/owner", bodyNoOwner, map[string]string{"Idempotency-Key": "k"}, true, 1, h.ReassignOwner)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Forbidden — 403", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h, mc := setupTestHandler(ctrl)
		mc.EXPECT().ReassignItemOwner(gomock.Any(), "c", int64(1), int64(5), gomock.Nil(), "k").Return(cartclient.ErrForbidden)
		w := doJSON(http.MethodPatch, "/api/cart/items/owner", bodyNoOwner, map[string]string{"Idempotency-Key": "k"}, true, 1, h.ReassignOwner)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Внутренняя ошибка — 500", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h, mc := setupTestHandler(ctrl)
		mc.EXPECT().ReassignItemOwner(gomock.Any(), "c", int64(1), int64(5), gomock.Nil(), "k").Return(errors.New("x"))
		w := doJSON(http.MethodPatch, "/api/cart/items/owner", bodyNoOwner, map[string]string{"Idempotency-Key": "k"}, true, 1, h.ReassignOwner)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestCartHandler_KickMember(t *testing.T) {
	body := KickMemberRequest{CartID: "c", TargetPublicID: "uuid-target"}

	t.Run("Неавторизован — 401", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h, _ := setupTestHandler(ctrl)
		w := doJSON(http.MethodDelete, "/api/cart/members", body, map[string]string{"Idempotency-Key": "k"}, false, 1, h.KickMember)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Без Idempotency-Key — 400", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h, _ := setupTestHandler(ctrl)
		w := doJSON(http.MethodDelete, "/api/cart/members", body, nil, true, 1, h.KickMember)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Битый JSON — 400", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h, _ := setupTestHandler(ctrl)
		w := doJSON(http.MethodDelete, "/api/cart/members", "bad", map[string]string{"Idempotency-Key": "k"}, true, 1, h.KickMember)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Целевой юзер не найден — 400", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		// Тут нам надо подменить userClient — у setupTestHandler у нас фолбэк-стаб возвращает ErrUserNotFound на UserNotFound,
		// но возвращает свою собственную ошибку. Подменим вручную:
		h, _ := setupTestHandler(ctrl)
		// fallthrough path: setupTestHandler настроил userClient на всегда-ошибку, но это не ErrUserNotFound,
		// поэтому handler выдаст 500. Этот сценарий проверяет именно «прочая ошибка ResolvePublicID — 500».
		w := doJSON(http.MethodDelete, "/api/cart/members", body, map[string]string{"Idempotency-Key": "k"}, true, 1, h.KickMember)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestCartHandler_KickMember_TargetNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cartMock := mocks.NewMockCartClient(ctrl)
	handler := makeHandlerForKickTest(t, ctrl, cartMock, userclient.ErrUserNotFound)

	body := KickMemberRequest{CartID: "c", TargetPublicID: "uuid-target"}
	w := doJSON(http.MethodDelete, "/api/cart/members", body, map[string]string{"Idempotency-Key": "k"}, true, 1, handler.KickMember)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCartHandler_KickMember_HappyAndForbidden(t *testing.T) {
	body := KickMemberRequest{CartID: "c", TargetPublicID: "uuid-target"}

	t.Run("Успех", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		cartMock := mocks.NewMockCartClient(ctrl)
		cartMock.EXPECT().KickMember(gomock.Any(), "c", int64(1), int64(42), "k").Return(nil)
		h := makeHandlerForKickTest(t, ctrl, cartMock, nil) // nil error => resolves to int64(42)
		w := doJSON(http.MethodDelete, "/api/cart/members", body, map[string]string{"Idempotency-Key": "k"}, true, 1, h.KickMember)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Forbidden — 403", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		cartMock := mocks.NewMockCartClient(ctrl)
		cartMock.EXPECT().KickMember(gomock.Any(), "c", int64(1), int64(42), "k").Return(cartclient.ErrForbidden)
		h := makeHandlerForKickTest(t, ctrl, cartMock, nil)
		w := doJSON(http.MethodDelete, "/api/cart/members", body, map[string]string{"Idempotency-Key": "k"}, true, 1, h.KickMember)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Внутренняя ошибка — 500", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		cartMock := mocks.NewMockCartClient(ctrl)
		cartMock.EXPECT().KickMember(gomock.Any(), "c", int64(1), int64(42), "k").Return(errors.New("boom"))
		h := makeHandlerForKickTest(t, ctrl, cartMock, nil)
		w := doJSON(http.MethodDelete, "/api/cart/members", body, map[string]string{"Idempotency-Key": "k"}, true, 1, h.KickMember)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestCartHandler_CloseSharedCart(t *testing.T) {
	body := BasicCartOperationRequest{CartID: "c"}

	t.Run("Неавторизован — 401", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h, _ := setupTestHandler(ctrl)
		w := doJSON(http.MethodPost, "/api/cart/close", body, map[string]string{"Idempotency-Key": "k"}, false, 1, h.CloseSharedCart)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("Без Idempotency-Key — 400", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h, _ := setupTestHandler(ctrl)
		w := doJSON(http.MethodPost, "/api/cart/close", body, nil, true, 1, h.CloseSharedCart)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("Битый JSON — 400", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h, _ := setupTestHandler(ctrl)
		w := doJSON(http.MethodPost, "/api/cart/close", "x", map[string]string{"Idempotency-Key": "k"}, true, 1, h.CloseSharedCart)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("Успех", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h, mc := setupTestHandler(ctrl)
		mc.EXPECT().CloseSharedCart(gomock.Any(), "c", int64(1), "k").Return(nil)
		w := doJSON(http.MethodPost, "/api/cart/close", body, map[string]string{"Idempotency-Key": "k"}, true, 1, h.CloseSharedCart)
		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("Forbidden — 403", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h, mc := setupTestHandler(ctrl)
		mc.EXPECT().CloseSharedCart(gomock.Any(), "c", int64(1), "k").Return(cartclient.ErrForbidden)
		w := doJSON(http.MethodPost, "/api/cart/close", body, map[string]string{"Idempotency-Key": "k"}, true, 1, h.CloseSharedCart)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("Внутренняя ошибка — 500", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h, mc := setupTestHandler(ctrl)
		mc.EXPECT().CloseSharedCart(gomock.Any(), "c", int64(1), "k").Return(errors.New("x"))
		w := doJSON(http.MethodPost, "/api/cart/close", body, map[string]string{"Idempotency-Key": "k"}, true, 1, h.CloseSharedCart)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestCartHandler_ClearCart_ErrorPaths(t *testing.T) {
	body := BasicCartOperationRequest{CartID: "c"}

	t.Run("Неавторизован — 401", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h, _ := setupTestHandler(ctrl)
		w := doJSON(http.MethodDelete, "/api/cart", body, map[string]string{"Idempotency-Key": "k"}, false, 1, h.ClearCart)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("Без Idempotency-Key — 400", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h, _ := setupTestHandler(ctrl)
		w := doJSON(http.MethodDelete, "/api/cart", body, nil, true, 1, h.ClearCart)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("Битый JSON — 400", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h, _ := setupTestHandler(ctrl)
		w := doJSON(http.MethodDelete, "/api/cart", "x", map[string]string{"Idempotency-Key": "k"}, true, 1, h.ClearCart)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("Forbidden — 403", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h, mc := setupTestHandler(ctrl)
		mc.EXPECT().ClearCart(gomock.Any(), "c", int64(1), "k").Return(cartclient.ErrForbidden)
		w := doJSON(http.MethodDelete, "/api/cart", body, map[string]string{"Idempotency-Key": "k"}, true, 1, h.ClearCart)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("Внутренняя ошибка — 500", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		h, mc := setupTestHandler(ctrl)
		mc.EXPECT().ClearCart(gomock.Any(), "c", int64(1), "k").Return(errors.New("x"))
		w := doJSON(http.MethodDelete, "/api/cart", body, map[string]string{"Idempotency-Key": "k"}, true, 1, h.ClearCart)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// Прямой контекст без middleware
func TestCartHandler_AllHandlers_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	h, _ := setupTestHandler(ctrl)

	cases := []struct {
		name string
		fn   func(http.ResponseWriter, *http.Request)
	}{
		{"AddItem", h.AddItem},
		{"GenerateInvite", h.GenerateInvite},
		{"RemoveItem", h.RemoveItem},
		{"UpdateQuantity", h.UpdateQuantity},
		{"ReassignOwner", h.ReassignOwner},
		{"JoinCart", h.JoinCart},
		{"KickMember", h.KickMember},
		{"CloseSharedCart", h.CloseSharedCart},
		{"ClearCart", h.ClearCart},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/whatever", bytes.NewBufferString(`{}`))
			w := httptest.NewRecorder()
			c.fn(w, req)
			// Подмотка userID не сделана — должен быть 401, кроме case-ов, где код доходит до парсинга авторизованным.
			// Все хэндлеры в этом списке первым делом проверяют middleware.GetUserID.
			assert.Equal(t, http.StatusUnauthorized, w.Code, "%s should require auth", c.name)
		})
	}

	// Контекст-проверка middleware.GetUserID без UserIDKey
	t.Run("GetUserID без ключа", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		_, ok := middleware.GetUserID(req.Context())
		assert.False(t, ok)
	})
}
