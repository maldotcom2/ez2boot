package middleware

import (
	"net/http"
)

// Set content type for all requests
func (m *Middleware) JsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set JSON Content-Type
		w.Header().Set("Content-Type", "application/json")
		// 1MB limit
		r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)
		next.ServeHTTP(w, r)
	})
}
