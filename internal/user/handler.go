package user

import (
	"encoding/json"
	"errors"
	"ez2boot/internal/contextkey"
	"ez2boot/internal/shared"
	"net/http"
	"time"
)

// UI endpoint for runtime state and bootstrap flow
func (h *Handler) CheckMode() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setupMode := h.Service.Config.SetupMode

		response := map[string]bool{
			"setup_mode": setupMode,
		}

		json.NewEncoder(w).Encode(response)
	}
}

func (h *Handler) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u UserLogin
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			h.Logger.Info("Malformed request", "email", u.Email, "error", err)
			http.Error(w, "Malformed request", http.StatusBadRequest)
			return
		}

		h.Logger.Info("Login attempted", "email", u.Email)

		token, err := h.Service.loginUser(u)
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

// Handler to create new user
func (h *Handler) CreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateUserRequest
		err := json.NewDecoder(r.Body).Decode(&req)

		h.Logger.Info("Attempted user creation", "email", req.Email)

		if err != nil {
			h.Logger.Error("Malformed request", "email", req.Email)
			http.Error(w, "Malformed request", http.StatusBadRequest)
			return
		}

		if err = h.Service.createUser(req); err != nil {
			h.Logger.Error("Failed to create user", "email", req.Email, "error", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

// Handler to bootstrap initial user creation - username and password input
func (h *Handler) CreateFirstTimeUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Block subsequent requests
		if h.Service.Config.SetupMode == false {
			h.Logger.Error("First time user creation blocked")
			http.Error(w, "First time user creation blocked", http.StatusForbidden)
			return
		}

		var req CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.Logger.Error("Malformed request", "email", req.Email)
			http.Error(w, "Malformed request", http.StatusBadRequest)
			return
		}

		h.Logger.Info("First time user creation", "email", req.Email)

		req.IsActive = true
		req.IsAdmin = true
		req.APIEnabled = true
		req.UIEnabled = true

		if err := h.Service.createUser(req); err != nil {
			h.Logger.Error("Failed to create user", "email", req.Email, "error", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		// Disable setup mode
		h.Service.Config.SetupMode = false

		w.WriteHeader(http.StatusCreated)
	}
}

func (h *Handler) ChangePassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(contextkey.UserIDKey).(int64)
		if !ok {
			h.Logger.Error("User ID not found in context")
			http.Error(w, "User ID not found in context", http.StatusUnauthorized)
			return
		}

		var c ChangePasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			h.Logger.Error("Malformed request", "userid", userID)
			http.Error(w, "Malformed request", http.StatusBadRequest)
			return
		}

		c.UserID = userID

		if err := h.Service.changePassword(c); err != nil {
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
	}
}
