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
			h.Logger.Info("Malformed request", "username", u.Username, "error", err)
			http.Error(w, "Malformed request", http.StatusBadRequest)
		}

		h.Logger.Info("Login attempted", "username", u.Username)

		token, err := h.Service.LoginUser(u)
		if err != nil {
			h.Logger.Info("Failed to login", "username", u.Username, "error", err)
			http.Error(w, "Failed to login", http.StatusInternalServerError)
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

		h.Logger.Info("User logged in", "username", u.Username)
	}
}

// Handler to register new user
func (h *Handler) RegisterUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u model.UserLogin
		err := json.NewDecoder(r.Body).Decode(&u)

		h.Logger.Info("Attempted registration", "username", u.Username)

		if err != nil {
			h.Logger.Info("Malformed request", "username", u.Username)
			http.Error(w, "Malformed request", http.StatusBadRequest)
			return
		}

		if err = h.Service.validateAndCreateUser(u); err != nil {
			h.Logger.Info("Failed to create user", "username", u.Username, "error", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func (h *Handler) ChangePassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var c model.ChangePasswordRequest
		err := json.NewDecoder(r.Body).Decode(&c)

		h.Logger.Info("Attempted password change", "username", c.Username)

		if err != nil {
			h.Logger.Error("Malformed request", "username", c.Username)
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
