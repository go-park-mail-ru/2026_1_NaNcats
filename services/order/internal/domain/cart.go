package domain

type CartItem struct {
	DishID      int64
	Name        string
	Quantity    int
	Price       int64
	OwnerUserID *int64
}

type Cart struct {
	RestaurantBrandID int64
	Items             []CartItem
	Members           []int64
}
