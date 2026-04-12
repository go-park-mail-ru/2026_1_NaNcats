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
)

func GetUserID(ctx context.Context) (int, error) {
	// берем userID из контекста, который нам пришел из мидлвара AuthMiddleware
	// Value возвращает any. Используем утверждение типа, чтобы Go знал что это uuid
	if id, ok := ctx.Value(UserIDKey).(int); ok {
		return id, nil
	}
	// если там не int или nil
	return 0, ErrNoUserIDInContext
}
