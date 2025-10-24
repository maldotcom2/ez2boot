package middleware

import (
	"context"
	"errors"
	"ez2boot/internal/repository"
	"ez2boot/internal/service/users"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
)

func BasicAuthMiddleware(repo *repository.Repository, logger *slog.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check basic auth password
			username, password, ok := r.BasicAuth()
			if !ok {
				logger.Warn("Unauthorised login attempt due to incorrect or missing auth header", "username", username, "source ip", r.RemoteAddr)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ok, err := users.ComparePassword(repo, username, password)
			if err != nil {
				logger.Error("Could not compare password for supplied user", "username", username, "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if !ok {
				logger.Warn("Unauthorised login attempt for user", "username", username, "source ip", r.RemoteAddr)
				w.WriteHeader(http.StatusUnauthorized)
				return
			} else {
				userID, err := users.GetBasicAuthInfo(repo, username)
				if err != nil {
					logger.Error("Could not retrieve userID for supplied basic auth user", "username", username, "source ip", r.RemoteAddr)
					return
				}

				logger.Info("Basic auth passed", "username", username, "Path", r.URL.Path, "Source IP", r.RemoteAddr)
				// Pass down request to the next middleware
				ctx := context.WithValue(r.Context(), UserIDKey, userID)
				next.ServeHTTP(w, r.WithContext(ctx))
			}
		})
	}
}

func SessionAuthMiddleware(repo *repository.Repository, logger *slog.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Check for session cookie
			cookie, err := r.Cookie("session_id")
			if err != nil || cookie.Value == "" {
				logger.Info("User didn't present a sessionID to middleware - client redirect", "Path", r.URL.Path, "Source IP", r.RemoteAddr)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Check whether the cookie is for a valid session, and get user account info
			u, err := users.GetSessionInfo(repo, cookie.Value)
			if err != nil {
				if errors.Is(err, users.ErrSessionNotFound) {
					logger.Info("Session not found for user", "error", err)
					http.Error(w, "Session not found", http.StatusUnauthorized)
					return
				}

				if errors.Is(err, users.ErrSessionExpired) {
					logger.Info("Session expired", "username", u.Username)
					http.Error(w, "Session expired", http.StatusUnauthorized)
					return
				}
			}

			// Create a context containing the userID and the account verified status. This controls the authorisation to downstream functions.
			logger.Info("User request passed middleware", "username", u.Username, "Path", r.URL.Path, "Source IP", r.RemoteAddr)
			ctx := context.WithValue(r.Context(), UserIDKey, u.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
