package domain

import (
	"time"

	"github.com/google/uuid"
)

// сущность сессии
type Session struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ExpiresAt time.Time
}
