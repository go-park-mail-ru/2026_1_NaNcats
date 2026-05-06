package domain

import "time"

type OrderSplit struct {
	ID                string
	OrderID           int64
	UserID            int64
	Amount            int64
	Status            string
	PaymentMethodID   *string
	YookassaPaymentID *string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
type Order struct {
	ID                 int64
	PublicID           string
	AdminID            int64
	CourierID          int64
	RestaurantBranchID int64
	RestaurantBrandID  int64
	ClientAddressID    string
	TotalCost          int64
	PromocodeID        int64
	RestaurantName     string
	Status             string
	Items              []OrderDish
	Splits             []OrderSplit
	CreatedAt          time.Time
	UpdatedAt          time.Time
	RestaurantLogoURL  string
}

type OrderDish struct {
	DishID      int64
	Quantity    int
	Price       int64
	OwnerUserID *int64
}

type CreateOrderInput struct {
	UserID             int64
	AddressPublicID    string
	RestaurantBranchID int64
	RestaurantBrandID  int64
	DeliveryCost       int64
	ServiceFee         int64

	PaymentMethodID string
	PayForAll       bool
	PayerMapping    map[int64]int64
}
