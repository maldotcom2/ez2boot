package shared

type ApiResponse[T any] struct {
	Success bool
	Data    T
	Error   string
}
