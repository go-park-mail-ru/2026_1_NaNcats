package domain

import "time"

// сущность сессии
type Session struct {
	ID        string
	UserID    int
	ExpiresAt time.Time
}
