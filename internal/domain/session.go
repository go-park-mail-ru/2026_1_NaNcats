package domain

//go:generate easyjson $GOFILE

import (
	"time"

	"github.com/google/uuid"
)

// сущность сессии
//
//easyjson:json
type Session struct {
	ID        uuid.UUID `json:"id"`
	UserID    int       `json:"user_id"`
	UserAgent string    `json:"user_agent"`
	Role      string    `json:"role"`
	ExpiresAt time.Time `json:"expires_at"`
}
