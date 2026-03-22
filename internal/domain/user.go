package domain

import (
	"time"
)

// сущность юзера
type User struct {
	ID           int
	Phone        string
	Name         string
	Email        string
	PasswordHash string
	AvatarURL    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
