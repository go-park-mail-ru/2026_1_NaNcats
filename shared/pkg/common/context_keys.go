package common

import "context"

type contextKey string

const RequestIDKey contextKey = "requestID"

const UserIDKey contextKey = "userID"

func GetUserID(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(UserIDKey).(int64)
	return id, ok
}
