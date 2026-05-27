package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPromocode_RestaurantBrandIDs(t *testing.T) {
	t.Run("Глобальный промо (без restaurant_brand_id) — пустой слайс", func(t *testing.T) {
		p := Promocode{IsGlobal: true, RestaurantBrandID: nil}
		got := p.RestaurantBrandIDs()
		assert.NotNil(t, got, "должен возвращать инициализированный пустой слайс, а не nil")
		assert.Empty(t, got)
	})

	t.Run("Промо для конкретного бренда — слайс из одного id", func(t *testing.T) {
		bid := int64(42)
		p := Promocode{RestaurantBrandID: &bid}
		got := p.RestaurantBrandIDs()
		assert.Equal(t, []int64{42}, got)
	})
}
