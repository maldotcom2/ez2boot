package ctxutil

// ContextKey is a custom type to avoid context key collisions
type ContextKey string

const (
	UserIDKey ContextKey = "userID"
	EmailKey  ContextKey = "email"
)
