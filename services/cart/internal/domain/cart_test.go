package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCart_HasMember(t *testing.T) {
	c := &Cart{
		Members: []CartMember{{UserID: 1}, {UserID: 2}, {UserID: 5}},
	}

	tests := []struct {
		name string
		uid  int64
		want bool
	}{
		{"Член есть в начале списка", 1, true},
		{"Член есть в середине", 2, true},
		{"Член есть в конце", 5, true},
		{"Не член — false", 999, false},
		{"Пустой ID — false", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, c.HasMember(tt.uid))
		})
	}

	t.Run("Пустая корзина — false", func(t *testing.T) {
		empty := &Cart{}
		assert.False(t, empty.HasMember(1))
	})
}

func TestCart_GetItem(t *testing.T) {
	uid1 := int64(1)
	uid2 := int64(2)
	c := &Cart{
		Items: []CartItem{
			{DishID: 10, OwnerUserID: &uid1, Quantity: 2, Name: "пицца"},
			{DishID: 10, OwnerUserID: &uid2, Quantity: 1, Name: "пицца"},
			{DishID: 20, OwnerUserID: nil, Quantity: 1, Name: "ничейная"},
		},
	}

	t.Run("Существующий dish/owner — возвращает указатель", func(t *testing.T) {
		got := c.GetItem(10, 1)
		assert.NotNil(t, got)
		assert.EqualValues(t, 10, got.DishID)
		assert.EqualValues(t, 2, got.Quantity)
	})

	t.Run("Тот же dish, другой owner — отдельная позиция", func(t *testing.T) {
		got := c.GetItem(10, 2)
		assert.NotNil(t, got)
		assert.EqualValues(t, 1, got.Quantity)
	})

	t.Run("Несуществующий dish — nil", func(t *testing.T) {
		assert.Nil(t, c.GetItem(99, 1))
	})

	t.Run("Существующий dish, но другой owner — nil", func(t *testing.T) {
		assert.Nil(t, c.GetItem(10, 999))
	})

	t.Run("Dish с nil-owner не возвращается при поиске по owner", func(t *testing.T) {
		assert.Nil(t, c.GetItem(20, 1))
	})
}
