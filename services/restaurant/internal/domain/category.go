package domain

import "time"

type Category struct {
	ID        int64
	Name      string
	Emoji     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
