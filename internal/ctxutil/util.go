package ctxutil

import "context"

func GetActor(ctx context.Context) (int64, string) {
	if uid, ok := GetUserID(ctx); ok {
		email, ok := GetEmail(ctx)
		if !ok {
			email = "system"
		}
		return uid, email
	}
	return 0, "system"
}
