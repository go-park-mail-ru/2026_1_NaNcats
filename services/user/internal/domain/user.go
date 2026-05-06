package domain

import (
	"time"
)

const (
	RoleClient  = "client"
	RoleCourier = "courier"
	RoleOwner   = "owner"
	RoleAdmin   = "admin"
	RoleSupport = "support"
)

// сущность юзера
type User struct {
	ID           int64
	Name         string
	Email        string
	PasswordHash string
	Role         string
	AvatarURL    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
