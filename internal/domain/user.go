package domain

import (
	"time"

	"github.com/google/uuid"
)

// сущность юзера
type User struct {
	ID           uuid.UUID
	Phone        string
	Name         string
	Email        string
	PasswordHash string
	AvatarURL    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
