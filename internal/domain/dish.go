package domain

import (
	"time"
)

// Блюдо/позиция ресторана
type Dish struct {
	ID                int
	RestaurantBrandID int
	Name              string
	Description       string
	ImageURL          string
	Price             int64 // BIGINT
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
