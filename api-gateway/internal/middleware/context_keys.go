package middleware

import "context"

// уникальный тип ключа контекста:
// нужен, чтобы предотвратить коллизии с другими пакетами в контексте
type contextKey string

const (
	// id реквеста для отслеживания в логах
	RequestIDKey contextKey = "requestID"
	// ключ ID пользователя для контекста
	UserIDKey contextKey = "userID"
	// ключ Роли пользователя (добавлен для полноты картины)
	RoleKey contextKey = "role"
)

func GetUserID(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(UserIDKey).(int64)
	return id, ok
}
