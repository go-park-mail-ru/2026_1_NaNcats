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
