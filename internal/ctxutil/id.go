package ctxutil

import "context"

func GetUserID(ctx context.Context) int64 {
	return ctx.Value(UserIDKey).(int64)
}
