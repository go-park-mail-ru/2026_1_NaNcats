package domain

import "time"

type Promocode struct {
	ID                int64
	Code              string
	DiscountPercent   *int
	DiscountAmount    *int64
	MaxUses           *int
	CurrentUses       int
	MinOrderAmount    *int64
	UserID            *int64
	RestaurantBrandID *int64
	IsGlobal          bool
	CreatedAt         time.Time
	ExpiresAt         time.Time
}

type PromocodeUsage struct {
	ID          int64
	PromocodeID int64
	OrderID     int64
	UserID      int64
	UsedAt      time.Time
}
