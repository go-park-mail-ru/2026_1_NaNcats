package events

const (
	QueueUserEvents = "user.events"

	// Типы событий
	EventTypeRoleChanged = "UserRoleChanged"
)

//easyjson:json
type UserRoleChangedEvent struct {
	UserID  int64  `json:"user_id"`
	OldRole string `json:"old_role"`
	NewRole string `json:"new_role"`
}
