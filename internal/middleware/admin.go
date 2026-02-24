package middleware

import (
	"encoding/json"
	"ez2boot/internal/ctxutil"
	"ez2boot/internal/shared"
	"net/http"
)

func (m *Middleware) AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, email := ctxutil.GetActor(ctx)

		user, err := m.UserService.GetUserAuthorisation(userID)
		if err != nil {
			m.Logger.Error("Failed to fetch user authorisation", "user", email, "path", r.URL.Path, "domain", "middleware", "error", err)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to fetch user authorisation"})
			return
		}

		if !user.IsAdmin {
			m.Logger.Warn("Non-admin user attempted to access admin functions", "user", email, "path", r.URL.Path, "domain", "middleware")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Unauthorised"})
			return
		}

		// Pass down request to the next middleware
		next.ServeHTTP(w, r)
	})
}
