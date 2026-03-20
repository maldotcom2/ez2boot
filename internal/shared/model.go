package shared

type ApiResponse[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data"`
	Error   string `json:"error"`
}

type AuthResult struct {
	UserID           int64
	IdentityProvider string
	Authenticated    bool
}

type UserCredentials struct {
	UserID           int64
	Email            string
	PasswordHash     *string // Can be null
	IdentityProvider string
}
