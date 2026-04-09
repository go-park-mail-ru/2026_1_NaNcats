package domain

import "time"

type ClientProfile struct {
	AccountID               int
	BonusBalance            int64
	BonusCategoryID         *int
	BonusCategoryExpiresAt  *time.Time
	BonusExpiresAt          *time.Time
	StreakCount             int
	LastOrderDate           *time.Time
	PremiumExpiresAt        *time.Time
}
