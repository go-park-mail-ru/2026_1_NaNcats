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
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
