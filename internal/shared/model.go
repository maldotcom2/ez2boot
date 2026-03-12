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

type UserInfo struct {
	UserID           int64
	PasswordHash     string
	IdentityProvider string
}
