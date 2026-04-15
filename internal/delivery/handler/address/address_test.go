package address

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	domainMocks "github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAddressHandler_AddAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := mocks.NewMockAddressUseCase(ctrl)
	mockLogger := domainMocks.NewNopLogger()
	h := NewAddressHandler(mockUseCase, mockLogger)

	addrReq := AddressRequest{
		AddressText: "Москва",
		Lat:         55.75,
		Lon:         37.61,
	}
	body, _ := json.Marshal(addrReq)

	t.Run("Успешное добавление адреса", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/address", bytes.NewBuffer(body))
		ctx := context.WithValue(r.Context(), middleware.UserIDKey, 1)
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		mockUseCase.EXPECT().
			AddAddress(gomock.Any(), 1, gomock.Any()).
			Return("uuid-123", nil)

		h.AddAddress(w, r)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "uuid-123")
	})

	t.Run("Ошибка декодирования JSON", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/address", bytes.NewBuffer([]byte("invalid")))
		ctx := context.WithValue(r.Context(), middleware.UserIDKey, 1)
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		h.AddAddress(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Ошибка в UseCase", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/address", bytes.NewBuffer(body))
		ctx := context.WithValue(r.Context(), middleware.UserIDKey, 1)
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		mockUseCase.EXPECT().
			AddAddress(gomock.Any(), 1, gomock.Any()).
			Return("", errors.New("db error"))

		h.AddAddress(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestAddressHandler_GetAddresses(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := mocks.NewMockAddressUseCase(ctrl)
	mockLogger := domainMocks.NewNopLogger()
	h := NewAddressHandler(mockUseCase, mockLogger)

	t.Run("Успешное получение списка адресов", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/address", nil)
		ctx := context.WithValue(r.Context(), middleware.UserIDKey, 1)
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		mockUseCase.EXPECT().
			GetMyAddresses(gomock.Any(), 1).
			Return([]domain.Address{{PublicID: "1"}}, nil)

		h.GetAddresses(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "addresses")
	})

	t.Run("Ошибка UseCase при получении", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/address", nil)
		ctx := context.WithValue(r.Context(), middleware.UserIDKey, 1)
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		mockUseCase.EXPECT().
			GetMyAddresses(gomock.Any(), 1).
			Return(nil, errors.New("error"))

		h.GetAddresses(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestAddressHandler_DeleteAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := mocks.NewMockAddressUseCase(ctrl)
	mockLogger := domainMocks.NewNopLogger()
	h := NewAddressHandler(mockUseCase, mockLogger)

	t.Run("Успешное удаление адреса", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodDelete, "/address/uuid-123", nil)
		r.SetPathValue("id", "uuid-123")
		ctx := context.WithValue(r.Context(), middleware.UserIDKey, 1)
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		mockUseCase.EXPECT().
			DeleteAddress(gomock.Any(), 1, "uuid-123").
			Return(nil)

		h.DeleteAddress(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Ошибка при удалении", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodDelete, "/address/uuid-123", nil)
		r.SetPathValue("id", "uuid-123")
		ctx := context.WithValue(r.Context(), middleware.UserIDKey, 1)
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		mockUseCase.EXPECT().
			DeleteAddress(gomock.Any(), 1, "uuid-123").
			Return(errors.New("fail"))

		h.DeleteAddress(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestAddressHandler_UpdateAddress(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUseCase := mocks.NewMockAddressUseCase(ctrl)
	mockLogger := domainMocks.NewNopLogger()
	h := NewAddressHandler(mockUseCase, mockLogger)

	addrReq := AddressRequest{AddressText: "Новый адрес"}
	body, _ := json.Marshal(addrReq)

	t.Run("Успешное обновление адреса", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPut, "/address/uuid-123", bytes.NewBuffer(body))
		r.SetPathValue("id", "uuid-123")
		ctx := context.WithValue(r.Context(), middleware.UserIDKey, 1)
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		mockUseCase.EXPECT().
			UpdateAddress(gomock.Any(), 1, gomock.Any()).
			Return(nil)

		h.UpdateAddress(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "successfully")
	})

	t.Run("Ошибка при обновлении", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPut, "/address/uuid-123", bytes.NewBuffer(body))
		r.SetPathValue("id", "uuid-123")
		ctx := context.WithValue(r.Context(), middleware.UserIDKey, 1)
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		mockUseCase.EXPECT().
			UpdateAddress(gomock.Any(), 1, gomock.Any()).
			Return(errors.New("error"))

		h.UpdateAddress(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
