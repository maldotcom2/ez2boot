package user

import (
	"encoding/json"
	"errors"
	"ez2boot/internal/contextkey"
	"ez2boot/internal/shared"
	"fmt"
	"net/http"
	"time"
)

func (h *Handler) GetUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := h.Service.getUsers()
		if err != nil {
			h.Logger.Error("Error while fetching users", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Error while fetching users"})
			return
		}

		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: users})
	}
}

func (h *Handler) UpdateUserAuthorisation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(contextkey.UserIDKey).(int64)
		if !ok {
			h.Logger.Error("User ID not found in context")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "User ID not found in context"})
			return
		}

		var req []UpdateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.Logger.Error("Malformed request", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Malformed request"})
			return
		}

		// Admin check
		if !h.userIsAdmin(w, r) {
			return // Response written by helper
		}

		if err := h.Service.updateUserAuthorisation(req, userID); err != nil {
			var resp shared.ApiResponse[any]
			switch {
			case errors.Is(err, shared.ErrCannotModifyOwnAuth):
				h.Logger.Error("User cannot modify their own authorisations")
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   err.Error(),
				}
			default:
				h.Logger.Info("Error updating user authorisation", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				resp = (shared.ApiResponse[any]{
					Success: false,
					Error:   "Error updating user authorisation"})
			}

			json.NewEncoder(w).Encode(resp)
			return
		}
	}
}

// UI endpoint for runtime state and bootstrap flow // TODO Does this belong here?
func (h *Handler) GetMode() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get mode from config
		setupMode := h.Service.Config.SetupMode

		response := SetupResponse{SetupMode: setupMode}

		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: response})
	}
}

// Session validity check for UI
func (h *Handler) CheckSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

// Get user authorisation for logged in user
func (h *Handler) GetUserAuthorisation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(contextkey.UserIDKey).(int64)
		if !ok {
			h.Logger.Error("User ID not found in context")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "User ID not found in context"})
			return
		}

		user, err := h.Service.GetUserAuthorisation(userID)
		if err != nil {
			h.Logger.Error("Error while fetching user authorisation")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Error while fetching user authorisation"})
			return
		}

		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: user})
	}
}

// Login session user
func (h *Handler) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u UserLogin
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			h.Logger.Error("Malformed request", "email", u.Email, "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Malformed request"})
			return
		}

		h.Logger.Info("Login attempted", "email", u.Email)

		var resp shared.ApiResponse[any]
		token, err := h.Service.login(u)
		if err != nil {
			switch {
			case errors.Is(err, shared.ErrEmailOrPasswordMissing):
				h.Logger.Error("Missing email or password for login")
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   shared.ErrEmailOrPasswordMissing.Error(),
				}
			case errors.Is(err, shared.ErrAuthenticationFailed):
				h.Logger.Error("Invalid email or password", "email", u.Email, "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   shared.ErrAuthenticationFailed.Error(),
				}
			default:
				h.Logger.Error("Failed to login", "email", u.Email, "error", err)
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
			SameSite: http.SameSiteStrictMode, // Use SameSiteLaxMode for testing
			HttpOnly: true,
			Secure:   true, // Use false for testing
		})

		h.Logger.Info("User logged in", "email", u.Email)
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}

// Logout session handler
func (h *Handler) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(contextkey.UserIDKey).(int64)
		if !ok {
			h.Logger.Error("User ID not found in context")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "User ID not found in context"})
			return
		}
		// Get session token
		cookie, _ := r.Cookie("session")

		email, err := h.Service.GetEmailFromUserID(userID)
		if err != nil {
			h.Logger.Error("Failed to get email from user id", "user id", userID, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Server error while processing logout"})
			return
		}

		if err := h.Service.logout(cookie.Value); err != nil {
			h.Logger.Error("Error while logging out user", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Server error while processing logout"})
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

		h.Logger.Info("User logged out", "email", email)
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}

// Handler to create new user
func (h *Handler) CreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.Logger.Error("Malformed request", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Malformed request"})
			return
		}

		h.Logger.Info("Attempted user creation", "email", req.Email)

		// Admin check
		if !h.userIsAdmin(w, r) {
			return // Response written by helper
		}

		var resp shared.ApiResponse[any]
		if err := h.Service.createUser(req); err != nil {
			switch {
			case errors.Is(err, shared.ErrUserAlreadyExists):
				h.Logger.Error("Failed to create user, user already exists")
				w.WriteHeader(http.StatusConflict)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   shared.ErrUserAlreadyExists.Error(),
				}
			case errors.Is(err, shared.ErrEmailPattern):
				h.Logger.Error("Failed to create user, email does not match pattern", "email", req.Email)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   shared.ErrEmailPattern.Error(),
				}
			case errors.Is(err, shared.ErrPasswordLength):
				h.Logger.Error("Failed to create user, password too short", "email", req.Email)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   shared.ErrPasswordLength.Error(),
				}
			case errors.Is(err, shared.ErrEmailContainsPassword):
				h.Logger.Error("Failed to create user, email contains password", "email", req.Email)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   shared.ErrEmailContainsPassword.Error(),
				}
			case errors.Is(err, shared.ErrPasswordContainsEmail):
				h.Logger.Error("Failed to create user, password contains email", "email", req.Email)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   shared.ErrPasswordContainsEmail.Error(),
				}
			default:
				h.Logger.Error("Failed to create user", "email", req.Email, "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   fmt.Sprintf("Failed to create user %s", err),
				}
			}

			json.NewEncoder(w).Encode(resp)
			return
		}

		h.Logger.Info("New user created", "email", req.Email)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}

func (h *Handler) DeleteUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(contextkey.UserIDKey).(int64)
		if !ok {
			h.Logger.Error("User ID not found in context")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "User ID not found in context"})
			return
		}

		var req DeleteUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.Logger.Error("Malformed request", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Malformed request"})
			return
		}

		// Admin check
		if !h.userIsAdmin(w, r) {
			return // Response written by helper
		}

		email, err := h.Service.GetEmailFromUserID(req.UserID)
		if err != nil {
			h.Logger.Error("Failed to get email from userID", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to get email from userID"})
		}

		var resp shared.ApiResponse[any]
		if err := h.Service.deleteUser(req.UserID, userID); err != nil {
			if errors.Is(err, shared.ErrCannotDeleteOwnUser) {
				h.Logger.Error("User attempted to delete own user", "email", email)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   shared.ErrCannotDeleteOwnUser.Error(),
				}
			} else {
				h.Logger.Error("Failed to delete user", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Failed to delete user",
				}
			}

			json.NewEncoder(w).Encode(resp)
			return
		}
	}
}

// Handler to bootstrap initial user creation - username and password input
func (h *Handler) CreateFirstTimeUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Block subsequent requests
		if h.Service.Config.SetupMode == false {
			h.Logger.Error("First time user creation blocked")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "First time user creation blocked"})
			return
		}

		var req CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.Logger.Error("Malformed request", "email", req.Email)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Malformed request"})
			return
		}

		h.Logger.Info("First time user creation", "email", req.Email)

		req.IsActive = true
		req.IsAdmin = true
		req.APIEnabled = true
		req.UIEnabled = true

		var resp shared.ApiResponse[any]
		if err := h.Service.createUser(req); err != nil {
			switch {
			case errors.Is(err, shared.ErrEmailPattern):
				h.Logger.Error("Failed to create user, email does not match pattern", "email", req.Email)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   shared.ErrEmailPattern.Error(),
				}
			case errors.Is(err, shared.ErrPasswordLength):
				h.Logger.Error("Failed to create user, password too short", "email", req.Email)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   shared.ErrPasswordLength.Error(),
				}
			case errors.Is(err, shared.ErrEmailContainsPassword):
				h.Logger.Error("Failed to create user, email contains password", "email", req.Email)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   shared.ErrEmailContainsPassword.Error(),
				}
			case errors.Is(err, shared.ErrPasswordContainsEmail):
				h.Logger.Error("Failed to create user, password contains email", "email", req.Email)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   shared.ErrPasswordContainsEmail.Error(),
				}
			default:
				h.Logger.Error("Failed to create user", "email", req.Email, "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   fmt.Sprintf("Failed to create user %s", err),
				}
			}
			json.NewEncoder(w).Encode(resp)
			return
		}

		// Disable setup mode
		h.Service.Config.SetupMode = false
		h.Logger.Info("First time user created", "email", req.Email)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}

func (h *Handler) ChangePassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(contextkey.UserIDKey).(int64)
		if !ok {
			h.Logger.Error("User ID not found in context")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "User ID not found in context"})
			return
		}

		var req ChangePasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.Logger.Error("Malformed request", "userid", userID)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Malformed request"})
			return
		}

		req.UserID = userID

		var resp shared.ApiResponse[any]
		email, err := h.Service.changePassword(req)
		if err != nil {
			switch {
			case errors.Is(err, shared.ErrOldOrNewPasswordMissing):
				h.Logger.Error("Failed to change password for user, old or new password missing", "email", email)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   shared.ErrOldOrNewPasswordMissing.Error(),
				}
			case errors.Is(err, shared.ErrAuthenticationFailed):
				h.Logger.Error("Failed to change password for user, authentication failed", "email", email)
				w.WriteHeader(http.StatusUnauthorized)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   shared.ErrAuthenticationFailed.Error(),
				}
			case errors.Is(err, shared.ErrInvalidPassword):
				h.Logger.Error("Failed to change password for user, password did not match complexity requirements", "email", email)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   shared.ErrInvalidPassword.Error(),
				}
			default:
				h.Logger.Error("Failed to change password for user", "email", email, "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Failed to change password",
				}

				json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to change password"})
				return
			}

			json.NewEncoder(w).Encode(resp)
			return
		}

		h.Logger.Info("Password changed for user", "email", email)
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}

// Handler helper to guard admin routes
func (h *Handler) userIsAdmin(w http.ResponseWriter, r *http.Request) bool {
	userID, ok := r.Context().Value(contextkey.UserIDKey).(int64)
	if !ok {
		h.Logger.Error("User ID not found in context")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "User ID not found in context"})
		return false
	}

	user, err := h.Service.GetUserAuthorisation(userID)
	if err != nil {
		h.Logger.Error("Error while fetching user authorisation")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Error while fetching user authorisation"})
		return false
	}

	if !user.IsAdmin {
		h.Logger.Error("Non-admin user attempted to access admin functions", "email", user.Email)
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Unauthorised"})
		return false
	}

	return true
}
