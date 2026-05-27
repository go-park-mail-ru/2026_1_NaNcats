package domain

import "time"

type ClientProfile struct {
	AccountID              int64
	BonusBalance           int64
	BonusCategoryID        *int64
	BonusCategoryExpiresAt *time.Time
	BonusExpiresAt         *time.Time
	StreakCount            int
	LastOrderDate          *time.Time
	PremiumExpiresAt       *time.Time
	StreakFreezeActive     bool
}

type WheelSpinResult struct {
	SectorID   int
	SectorName string
	Emoji      string
	PromoCode  *string
	ExpiresAt  *string
	Message    string
}

type WheelSector struct {
	ID     int
	Name   string
	Emoji  string
	Weight int
}
