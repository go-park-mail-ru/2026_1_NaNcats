package domain

import (
	"time"
)

// Сущность ресторанного бренда
type RestaurantBrand struct {
	ID             int64
	OwnerProfileID int64
	Name           string
	Description    string
	PromotionTier  int
	LogoURL        string
	// BannerURL — широкая обложка ресторана для баннера на странице; пусто, если не задана.
	BannerURL string
	CreatedAt time.Time
	UpdatedAt time.Time
}
