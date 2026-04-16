package common

import "context"

type contextKey string

const RequestIDKey contextKey = "requestID"

const UserIDKey contextKey = "userID"

func GetUserID(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(UserIDKey).(int)
	return id, ok
}
