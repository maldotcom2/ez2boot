package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"ez2boot/internal/ctxutil"
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
				json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false})
				return
			}

			userID, authenticated, err := m.UserService.AuthenticateUser(email, password)
			if err != nil {
				if errors.Is(err, shared.ErrUserNotFound) {
					m.Logger.Warn("Attempted login for user which does not exist", "email", email, "error", err)
					w.WriteHeader(http.StatusUnauthorized)
					json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Password incorrect or user not found"})
					return
				}
				m.Logger.Error("Could not compare password for supplied user", "email", email, "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Error logging in"})
				return
			}

			if !authenticated {
				m.Logger.Warn("Basic auth login attempt failed", "email", email, "source ip", r.RemoteAddr)
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Password incorrect or user not found"})
				return
			}

			// Get user permissions
			u, err := m.UserService.GetUserAuthorisation(userID)
			if err != nil {
				m.Logger.Error("Error while fetching user authorisation", "email", email, "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Error while fetching user authorisation"})
				return
			}

			if !u.IsActive {
				m.Logger.Info("Inactive user attempted login", "email", u.Email)
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "User is not active"})
				return
			}

			if !u.APIEnabled {
				m.Logger.Info("Non-API user attempted to reach API endpoint", "email", u.Email)
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "User not authorised for API access"})
				return
			}

			m.Logger.Info("Basic auth passed", "userID", u.UserID, "email", u.Email, "Path", r.URL.Path, "Source IP", r.RemoteAddr)
			// Pass down request to the next middleware
			ctx := context.WithValue(r.Context(), ctxutil.UserIDKey, u.UserID)
			ctx = context.WithValue(ctx, ctxutil.EmailKey, u.Email)
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
				json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false})
				return
			}

			// Check whether the cookie is for a valid session
			us, err := m.UserService.GetSessionStatus(cookie.Value)
			if err != nil {
				if errors.Is(err, shared.ErrSessionNotFound) {
					m.Logger.Info("User session not found for user", "error", err)
					w.WriteHeader(http.StatusUnauthorized)
					json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "User session not found"})
					return
				}

				if errors.Is(err, shared.ErrSessionExpired) {
					m.Logger.Info("User session expired", "email", us.Email)
					w.WriteHeader(http.StatusUnauthorized)
					json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "User session expired"})
					return
				}

				m.Logger.Error("An error occured while evaluating user session", "email", us.Email, "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Error while evaluating user session"})
				return
			}

			// Get user permissions
			ua, err := m.UserService.GetUserAuthorisation(us.UserID)
			if err != nil {
				m.Logger.Error("Error while fetching user authorisation", "email", us.Email, "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Error while fetching user authorisation"})
				return
			}

			if !ua.IsActive {
				m.Logger.Warn("Inactive user attempted login", "email", ua.Email)
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "User is not active"})
				return
			}

			if !ua.UIEnabled {
				m.Logger.Warn("Non-UI user attempted to login via UI", "email", ua.Email)
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "User not authorised for UI access"})
				return
			}

			// Create a context containing the userID and the account verified status. This controls the authorisation to downstream functions.
			m.Logger.Info("User request passed middleware", "userID", ua.UserID, "email", ua.Email, "Path", r.URL.Path, "Source IP", r.RemoteAddr)
			ctx := context.WithValue(r.Context(), ctxutil.UserIDKey, ua.UserID)
			ctx = context.WithValue(ctx, ctxutil.EmailKey, ua.Email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
