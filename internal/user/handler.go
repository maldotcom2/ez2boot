package user

import (
	"encoding/json"
	"errors"
	"ez2boot/internal/ctxutil"
	"ez2boot/internal/shared"
	"fmt"
	"net/http"
	"time"
)

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
				h.Logger.Warn("Login failed", "reason", shared.ErrEmailOrPasswordMissing)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Missing email or password for login",
				}
			case errors.Is(err, shared.ErrAuthenticationFailed):
				h.Logger.Warn("Login failed", "email", u.Email, "reason", shared.ErrAuthenticationFailed, "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Invalid email or password",
				}
			case errors.Is(err, shared.ErrUserInactive):
				h.Logger.Warn("Login failed", "email", u.Email, "reason", shared.ErrUserInactive, "error", err)
				w.WriteHeader(http.StatusForbidden)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "User inactive",
				}
			case errors.Is(err, shared.ErrUserNotAuthorised):
				h.Logger.Warn("User inactive", "email", u.Email, "reason", shared.ErrUserNotAuthorised, "error", err)
				w.WriteHeader(http.StatusForbidden)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "User not authorised",
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
			SameSite: h.Config.SameSiteMode, // Use SameSiteLaxMode for testing
			HttpOnly: true,
			Secure:   h.Config.SecureCookie, // Use false for testing
		})

		h.Logger.Info("User logged in", "email", u.Email)
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

// Get user authorisation for logged in user
func (h *Handler) GetUserAuthorisation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, _ := ctxutil.GetActor(ctx)

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

func (h *Handler) UpdateUserAuthorisation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req []UpdateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.Logger.Error("Malformed request", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Malformed request"})
			return
		}

		// Admin check
		if !h.UserIsAdmin(w, r) {
			return // Response written by helper
		}

		if err := h.Service.updateUserAuthorisation(req, ctx); err != nil {
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

// Handler to create new user
func (h *Handler) CreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.Logger.Error("Malformed request", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Malformed request"})
			return
		}

		h.Logger.Info("Attempted user creation", "email", req.Email)

		// Admin check
		if !h.UserIsAdmin(w, r) {
			return // Response written by helper
		}

		var resp shared.ApiResponse[any]
		if err := h.Service.createUser(req, ctx); err != nil {
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
		ctx := r.Context()

		var req DeleteUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.Logger.Error("Malformed request", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Malformed request"})
			return
		}

		// Admin check
		if !h.UserIsAdmin(w, r) {
			return // Response written by helper
		}

		email, err := h.Service.GetEmailFromUserID(req.UserID)
		if err != nil {
			h.Logger.Error("Failed to get email from userID", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to get email from userID"})
		}

		var resp shared.ApiResponse[any]
		if err := h.Service.deleteUser(req.UserID, ctx); err != nil {
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
		ctx := r.Context()
		// Block subsequent requests
		if !h.Service.Config.SetupMode {
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
		if err := h.Service.createUser(req, ctx); err != nil {
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
		ctx := r.Context()
		userID, email := ctxutil.GetActor(ctx)

		var req ChangePasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.Logger.Error("Malformed request", "userid", userID)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Malformed request"})
			return
		}

		var resp shared.ApiResponse[any]
		err := h.Service.changePassword(req, ctx)
		if err != nil {
			switch {
			case errors.Is(err, shared.ErrCurrentOrNewPasswordMissing):
				h.Logger.Error("Failed to change password for user, current or new password missing", "email", email)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "current or new password missing",
				}
			case errors.Is(err, shared.ErrAuthenticationFailed):
				h.Logger.Error("Failed to change password for user, authentication failed", "email", email)
				w.WriteHeader(http.StatusUnauthorized)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "authentication failed",
				}
			case errors.Is(err, shared.ErrInvalidPassword):
				h.Logger.Error("Failed to change password for user, password did not match complexity requirements", "email", email)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "password did not match complexity requirements",
				}
			case errors.Is(err, shared.ErrNoRowsUpdated):
				h.Logger.Error("Failed to change password for user, no rows were updated", "email", email)
				w.WriteHeader(http.StatusInternalServerError)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "no rows updated",
				}
			case errors.Is(err, shared.ErrNoRowsDeleted):
				h.Logger.Warn("Failed to clear sessions after password change, no rows were deleted", "email", email)
				// Not an actual error
				resp = shared.ApiResponse[any]{
					Success: true,
				}
			default:
				h.Logger.Error("Failed to change password for user", "email", email, "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Failed to change password",
				}
			}

			json.NewEncoder(w).Encode(resp)
			return
		}

		h.Logger.Info("Password changed for user", "email", email)
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}

// Handler helper to guard admin routes
func (h *Handler) UserIsAdmin(w http.ResponseWriter, r *http.Request) bool {
	ctx := r.Context()
	userID, _ := ctxutil.GetActor(ctx)

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
