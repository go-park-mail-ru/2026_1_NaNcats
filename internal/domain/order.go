package domain

import "time"

type Order struct {
	ID                 int
	PublicID           string
	ClientID           int
	CourierID          *int
	RestaurantBranchID int
	ClientAddressID    int
	TotalCost          int64
	Status             string
	PaymentMethodID    string
	YookassaPaymentID  string
	Items              []OrderDish
	CreatedAt          time.Time
}

type OrderDish struct {
	DishID   int
	Quantity int
	Price    int64
}

type CreateOrderInput struct {
	UserID             int
	AddressPublicID    string
	RestaurantBranchID int
	PaymentMethodID    string
}
