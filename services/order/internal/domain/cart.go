package domain

type CartItem struct {
	DishID   int64
	Quantity int
	Price    int64
}

type Cart struct {
	RestaurantBrandID int64
	Items             []CartItem
}
