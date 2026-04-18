package domain

import "time"

type Order struct {
	ID                 int64
	PublicID           string
	ClientID           int64
	CourierID          *int64
	RestaurantBranchID int
	ClientAddressID    int64
	TotalCost          int64
	Status             string
	PaymentMethodID    string
	YookassaPaymentID  string
	Items              []OrderDish
	CreatedAt          time.Time
	RestaurantName     string
	RestaurantLogoURL  string // удалить
}

type OrderDish struct {
	DishID   int64
	Quantity int
	Price    int64
}

type CreateOrderInput struct {
	UserID             int64
	AddressPublicID    string
	RestaurantBranchID int64
	PaymentMethodID    string
	DeliveryCost       int64
	ServiceFee         int64
	TotalCost          int64
}
