package domain

import "time"

const (
	AchievementCodeFirstOrder    = "first_order"
	AchievementCodeFiveOrders    = "five_orders"
	AchievementCodeGourmandThree = "gourmand_three"
	AchievementCodeStreakSixth   = "streak_six"
	AchievementFirstSpin         = "first_spin"
)

type Achievement struct {
	ID          int64
	Code        string
	Title       string
	Description string
	Icon        string
	SortOrder   int
}

type UserAchievement struct {
	AchievementID int64
	AwardedAt     time.Time
}
