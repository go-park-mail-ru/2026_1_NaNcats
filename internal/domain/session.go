package domain

import "time"

type Session struct {
	ID        string
	UserID    int
	ExpiresAt time.Time
}
