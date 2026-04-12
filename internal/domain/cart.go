package domain

import "time"

type CartItem struct {
	DishID   int
	Quantity int
	Name     string
	Price    int64
	ImageURL string
}

type Cart struct {
	Items             []CartItem
	UserID            int
	RestaurantBrandID int
	UpdatedAt         time.Time
}
