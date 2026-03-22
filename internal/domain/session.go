package domain

import (
	"time"

	"github.com/google/uuid"
)

// сущность сессии
type Session struct {
	ID        uuid.UUID
	UserID    int
	ExpiresAt time.Time
}
