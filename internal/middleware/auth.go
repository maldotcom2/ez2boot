package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
)

func AuthMiddleware(logger *slog.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check basic auth password
			username, password, ok := r.BasicAuth()
			if !ok {
				logger.Warn("Unauthorised login attempt due to incorrect or missing auth header", "username", username, "source ip", r.RemoteAddr)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			// TODO implement actual auth check
			logger.Info("Basic auth passed", "username", username, "password", password)
			// Pass down request to the next middleware
			next.ServeHTTP(w, r)
		})
	}
}
