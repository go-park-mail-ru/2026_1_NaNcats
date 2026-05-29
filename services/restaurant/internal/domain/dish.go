package domain

import (
	"time"
)

// Блюдо/позиция ресторана
type Dish struct {
	ID                int64
	RestaurantBrandID int64
	Name              string
	Description       string
	ImageURL          string
	Price             int64 // BIGINT
	// Section — раздел меню ресторана (напр. "Супы", "Десерты"); пустая строка, если не задан.
	Section   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
