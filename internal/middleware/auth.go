package middleware

import (
	"context"
	"errors"
	"ez2boot/internal/shared"
	"net/http"

	"github.com/gorilla/mux"
)

func (m *Middleware) BasicAuthMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check basic auth password
			email, password, ok := r.BasicAuth()
			if !ok {
				m.Logger.Warn("Unauthorised login attempt due to incorrect or missing auth header", "email", email, "source ip", r.RemoteAddr)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			_, ok, err := m.UserService.AuthenticateUser(email, password)
			if err != nil {
				if errors.Is(err, shared.ErrUserNotFound) {
					m.Logger.Warn("Attempted login for user which does not exist", "email", email, "error", err)
					w.WriteHeader(http.StatusUnauthorized) // Keep vague to avoid enumeration
					return
				}
				m.Logger.Error("Could not compare password for supplied user", "email", email, "error", err)
				http.Error(w, "Unable to login", http.StatusInternalServerError)
				return
			}

			if !ok {
				m.Logger.Warn("Unauthorised login attempt for user", "email", email, "source ip", r.RemoteAddr)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Get user permissions
			u, err := m.UserService.GetUserAuthorisation(email)
			if err != nil {
				m.Logger.Error("Error while fetching user authorisation", "email", email, "error", err)
				http.Error(w, "Error while fetching user authorisation", http.StatusInternalServerError)
				return
			}

			if !u.IsActive {
				m.Logger.Info("Inactive user attempted login", "email", u.Email)
				http.Error(w, "User is not active", http.StatusForbidden)
				return
			}

			if !u.APIEnabled {
				m.Logger.Info("Non-API user attempted to reach API endpoint", "email", u.Email)
				http.Error(w, "User not authorised for API access", http.StatusForbidden)
				return
			}

			m.Logger.Info("Basic auth passed", "email", email, "Path", r.URL.Path, "Source IP", r.RemoteAddr)
			// Pass down request to the next middleware
			ctx := context.WithValue(r.Context(), UserIDKey, u.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (m *Middleware) SessionAuthMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Check for session cookie
			cookie, err := r.Cookie("session")
			if err != nil || cookie.Value == "" {
				m.Logger.Info("User didn't present a sessionID to middleware - client redirect", "Path", r.URL.Path, "Source IP", r.RemoteAddr)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Check whether the cookie is for a valid session
			us, err := m.UserService.GetSessionStatus(cookie.Value)
			if err != nil {
				if errors.Is(err, shared.ErrSessionNotFound) {
					m.Logger.Info("Session not found for user", "error", err)
					http.Error(w, "Session not found", http.StatusUnauthorized)
					return
				}

				if errors.Is(err, shared.ErrSessionExpired) {
					m.Logger.Info("Session expired", "email", us.Email)
					http.Error(w, "Session expired", http.StatusUnauthorized)
					return
				}

				m.Logger.Error("An error occured while evaluating session", "email", us.Email, "error", err)
				http.Error(w, "Error while evaluating session", http.StatusInternalServerError)
				return
			}

			// Get user permissions
			ua, err := m.UserService.GetUserAuthorisation(us.Email)
			if err != nil {
				m.Logger.Error("Error while fetching user authorisation", "email", ua.Email, "error", err)
				http.Error(w, "Error while fetching user authorisation", http.StatusInternalServerError)
				return
			}

			if !ua.IsActive {
				m.Logger.Info("Inactive user attempted login", "email", ua.Email)
				http.Error(w, "User is not active", http.StatusForbidden)
				return
			}

			if !ua.UIEnabled {
				m.Logger.Info("Non-UI user attempted to login via UI", "email", ua.Email)
				http.Error(w, "User not authorised for UI access", http.StatusForbidden)
				return
			}

			// Create a context containing the userID and the account verified status. This controls the authorisation to downstream functions.
			m.Logger.Info("User request passed middleware", "email", ua.Email, "Path", r.URL.Path, "Source IP", r.RemoteAddr)
			ctx := context.WithValue(r.Context(), UserIDKey, ua.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
