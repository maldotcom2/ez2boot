package user

import (
	"encoding/json"
	"errors"
	"ez2boot/internal/model"
	"ez2boot/internal/shared"
	"net/http"
	"time"
)

func (h *Handler) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u model.UserLogin
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			h.Logger.Info("Malformed request", "email", u.Email, "error", err)
			http.Error(w, "Malformed request", http.StatusBadRequest)
			return
		}

		h.Logger.Info("Login attempted", "email", u.Email)

		token, err := h.Service.LoginUser(u)
		if err != nil {
			if errors.Is(err, shared.ErrAuthenticationFailed) {
				h.Logger.Info("Invalid email or password", "email", u.Email, "error", err)
				http.Error(w, "Invalid email or password", http.StatusUnauthorized)
				return
			}
			h.Logger.Info("Failed to login", "email", u.Email, "error", err)
			http.Error(w, "Failed to login", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    token,
			Path:     "/",
			Expires:  time.Now().Add(h.Service.Config.UserSessionDuration),
			SameSite: http.SameSiteNoneMode,
			HttpOnly: true,
			Secure:   true,
		})

		h.Logger.Info("User logged in", "email", u.Email)
	}
}

func (h *Handler) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session token
		cookie, _ := r.Cookie("session")

		if err := h.Service.logoutUser(cookie.Value); err != nil {
			h.Logger.Error("Error while logging out user", "error", err)
			http.Error(w, "Failed to logout", http.StatusInternalServerError)
			// TODO Redirect to login page
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

		h.Logger.Info("User logged out")
		// TODO redirect
	}
}

// Handler to register new user
func (h *Handler) RegisterUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u model.UserLogin
		err := json.NewDecoder(r.Body).Decode(&u)

		h.Logger.Info("Attempted registration", "email", u.Email)

		if err != nil {
			h.Logger.Info("Malformed request", "email", u.Email)
			http.Error(w, "Malformed request", http.StatusBadRequest)
			return
		}

		if err = h.Service.validateAndCreateUser(u); err != nil {
			h.Logger.Info("Failed to create user", "email", u.Email, "error", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func (h *Handler) ChangePassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var c model.ChangePasswordRequest
		err := json.NewDecoder(r.Body).Decode(&c)

		h.Logger.Info("Attempted password change", "email", c.Email)

		if err != nil {
			h.Logger.Error("Malformed request", "email", c.Email)
			http.Error(w, "Malformed request", http.StatusBadRequest)
			return
		}

		// TODO add sentinel error context
		if err = h.Service.changePasswordByUser(c); err != nil {
			if errors.Is(err, shared.ErrAuthenticationFailed) {
				h.Logger.Error("Failed to change password for user")
				http.Error(w, "Authentication failed", http.StatusUnauthorized)
				return
			} else if errors.Is(err, shared.ErrInvalidPassword) {
				h.Logger.Error("Failed to change password for user")
				http.Error(w, "Password did not match complexity requirements", http.StatusBadRequest)
				return
			} else {
				h.Logger.Error("Failed to change password for user")
				http.Error(w, "Failed to change password", http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
