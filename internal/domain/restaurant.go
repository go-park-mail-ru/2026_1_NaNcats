package domain

import (
	"time"
)

// Сущность ресторанного бренда
type RestaurantBrand struct {
	ID            int
	Name          string
	Description   string
	PromotionTier int
	LogoURL       string
	BannerURL     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
