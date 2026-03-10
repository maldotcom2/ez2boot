package auth

import (
	"encoding/json"
	"errors"
	"ez2boot/internal/ctxutil"
	"ez2boot/internal/shared"
	"net/http"
	"time"
)

func (h *Handler) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u UserLogin
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			h.Logger.Error("Malformed request", "user", u.Email, "domain", "auth", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Malformed request"})
			return
		}

		h.Logger.Debug("Login attempted", "user", u.Email, "domain", "auth")

		var resp shared.ApiResponse[any]
		token, err := h.Service.login(u)
		if err != nil {
			switch {
			case errors.Is(err, shared.ErrEmailOrPasswordMissing):
				h.Logger.Warn("Login failed", "user", u.Email, "domain", "auth", "error", err)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Missing email or password for login",
				}
			case errors.Is(err, shared.ErrUserNotFound):
				h.Logger.Warn("Login failed", "user", u.Email, "domain", "auth", "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Invalid email or password", // Make sure this stays the same as for auth fail
				}
			case errors.Is(err, shared.ErrAuthenticationFailed):
				h.Logger.Warn("Login failed", "user", u.Email, "domain", "auth", "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Invalid email or password", // Make sure this stays the same as for user not found
				}
			case errors.Is(err, shared.ErrUserInactive):
				h.Logger.Warn("Login failed", "user", u.Email, "domain", "auth", "error", err)
				w.WriteHeader(http.StatusForbidden)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "User not authorised",
				}
			case errors.Is(err, shared.ErrUserNotAuthorised):
				h.Logger.Warn("Login failed", "user", u.Email, "domain", "auth", "error", err)
				w.WriteHeader(http.StatusForbidden)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "User not authorised",
				}
			default:
				h.Logger.Error("Failed to login", "user", u.Email, "domain", "auth", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Failed to login",
				}
			}

			json.NewEncoder(w).Encode(resp)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    token,
			Path:     "/",
			Expires:  time.Now().Add(h.Service.Config.UserSessionDuration),
			SameSite: h.Config.SameSiteMode,
			HttpOnly: true,
			Secure:   h.Config.SecureCookie,
		})

		h.Logger.Debug("User logged in", "user", u.Email, "domain", "auth")
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}

// Logout session handler
func (h *Handler) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, email := ctxutil.GetActor(ctx)

		// Get session token
		cookie, _ := r.Cookie("session")

		if err := h.Service.logout(cookie.Value, ctx); err != nil {
			h.Logger.Error("Failed to logout user", "user", email, "domain", "auth", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to logout user"})
			return
		}

		// Expire and null cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			MaxAge:   -1,
		})

		h.Logger.Debug("User logged out", "user", email, "domain", "auth")
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}
