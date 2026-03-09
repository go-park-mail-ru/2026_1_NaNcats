package domain

import (
	"time"

	"github.com/google/uuid"
)

// Сущность ресторанного бренда
type RestaurantBrand struct {
	ID            uuid.UUID
	Name          string
	Description   string
	PromotionTier int
	LogoURL       string
	BannerURL     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
