package domain

import "time"

type CartItem struct {
	DishID   int64
	Quantity int
	Name     string
	Price    int64
	ImageURL string
}

type Cart struct {
	Items             []CartItem
	UserID            int64
	RestaurantBrandID int64
	UpdatedAt         time.Time
}
