package ctxutil

import "context"

func GetEmail(ctx context.Context) string {
	return ctx.Value(EmailKey).(string)
}
