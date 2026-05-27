package events

import "time"

const (
	QueueUserEvents = "user.order_paid"

	// Типы событий
	EventTypeRoleChanged = "UserRoleChanged"
	EventTypeOrderPaid   = "OrderPaid"
)

//easyjson:json
type UserRoleChangedEvent struct {
	UserID  int64  `json:"user_id"`
	OldRole string `json:"old_role"`
	NewRole string `json:"new_role"`
}

//easyjson:json
type OrderPaidEvent struct {
	UserID        int64     `json:"user_id"`
	RestaurantID  int64     `json:"restaurant_id"`
	OrderPublicID string    `json:"order_public_id"`
	PaidAt        time.Time `json:"paid_at"`
}
