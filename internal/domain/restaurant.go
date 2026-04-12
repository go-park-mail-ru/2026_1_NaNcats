package domain

import (
	"time"
)

// Сущность ресторанного бренда
type RestaurantBrand struct {
	ID             int
	OwnerProfileID int
	Name           string
	Description    string
	PromotionTier  int
	LogoURL        string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
