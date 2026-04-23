package domain

import "time"

type Order struct {
	ID                 int64
	PublicID           string
	ClientID           int64
	CourierID          int64
	RestaurantBranchID int64
	RestaurantBrandID  int64
	ClientAddressID    string
	TotalCost          int64
	PromocodeID        int64
	RestaurantName     string
	PaymentMethodID    string
	YookassaPaymentID  string
	Status             string
	Items              []OrderDish
	CreatedAt          time.Time
	UpdatedAt          time.Time
	RestaurantLogoURL  string
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
	RestaurantBrandID  int64
	PaymentMethodID    string
	DeliveryCost       int64
	ServiceFee         int64
}
