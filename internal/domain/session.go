package domain

import (
	"time"

	"github.com/google/uuid"
)

// сущность сессии
type Session struct {
	ID        uuid.UUID
	UserID    int
	UserAgent string
	Role      string
	ExpiresAt time.Time
}
