package user

import (
	"encoding/base64"
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
			h.Logger.Error("Malformed request", "user", u.Email, "domain", "user", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Malformed request"})
			return
		}

		h.Logger.Debug("Login attempted", "user", u.Email, "domain", "user")

		var resp shared.ApiResponse[any]
		token, err := h.Service.login(u)
		if err != nil {
			switch {
			case errors.Is(err, shared.ErrEmailOrPasswordMissing):
				h.Logger.Warn("Login failed", "user", u.Email, "domain", "user", "error", err)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Missing email or password for login",
				}
			case errors.Is(err, shared.ErrUserNotFound):
				h.Logger.Warn("Login failed", "user", u.Email, "domain", "user", "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Invalid email or password", // Make sure this stays the same as for auth fail
				}
			case errors.Is(err, shared.ErrAuthenticationFailed):
				h.Logger.Warn("Login failed", "user", u.Email, "domain", "user", "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Invalid email or password", // Make sure this stays the same as for user not found
				}
			case errors.Is(err, shared.ErrUserInactive):
				h.Logger.Warn("Login failed", "user", u.Email, "domain", "user", "error", err)
				w.WriteHeader(http.StatusForbidden)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "User not authorised",
				}
			case errors.Is(err, shared.ErrUserNotAuthorised):
				h.Logger.Warn("Login failed", "user", u.Email, "domain", "user", "error", err)
				w.WriteHeader(http.StatusForbidden)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "User not authorised",
				}
			default:
				h.Logger.Error("Failed to login", "user", u.Email, "domain", "user", "error", err)
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

		h.Logger.Debug("User logged in", "user", u.Email, "domain", "user")
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
			h.Logger.Error("Failed to logout user", "user", email, "domain", "user", "error", err)
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

		h.Logger.Debug("User logged out", "user", email, "domain", "user")
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}

func (h *Handler) GetUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, email := ctxutil.GetActor(ctx)

		users, err := h.Service.getUsers()
		if err != nil {
			h.Logger.Error("Failed to fetch users", "user", email, "domain", "user", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to fetch users"})
			return
		}

		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: users})
	}
}

// Get user authorisation for logged in user
func (h *Handler) GetUserAuthorisation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, email := ctxutil.GetActor(ctx)

		user, err := h.Service.GetUserAuthorisation(userID)
		if err != nil {
			h.Logger.Error("Failed to fetch user authorisation", "user", email, "domain", "user", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to fetch user authorisation"})
			return
		}

		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: user})
	}
}

func (h *Handler) UpdateUserAuthorisation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, email := ctxutil.GetActor(ctx)

		var req []UpdateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.Logger.Error("Malformed request", "user", email, "domain", "user", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Malformed request"})
			return
		}

		if err := h.Service.updateUserAuthorisation(req, ctx); err != nil {
			var resp shared.ApiResponse[any]
			switch {
			case errors.Is(err, shared.ErrCannotModifyOwnAuth):
				h.Logger.Error("Failed", "user", email, "domain", "user", "error", err)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Failed to update user authorisation",
				}
			default:
				h.Logger.Error("Failed to update user authorisation", "user", email, "domain", "user", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				resp = (shared.ApiResponse[any]{
					Success: false,
					Error:   "Failed to update user authorisation"})
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
		_, email := ctxutil.GetActor(ctx)

		var req CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.Logger.Error("Malformed request", "user", email, "domain", "user", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Malformed request"})
			return
		}

		h.Logger.Info("Attempted user creation", "email", req.Email)

		var resp shared.ApiResponse[any]
		if err := h.Service.createUser(req, ctx); err != nil {
			switch {
			case errors.Is(err, shared.ErrUserAlreadyExists):
				h.Logger.Error("Failed to create user", "user", email, "domain", "user", "target_user", req.Email, "error", err)
				w.WriteHeader(http.StatusConflict)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "User already exists",
				}
			case errors.Is(err, shared.ErrEmailPattern):
				h.Logger.Error("Failed to create user", "user", email, "domain", "user", "target_user", req.Email, "error", err)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Invalid email",
				}
			case errors.Is(err, shared.ErrPasswordLength):
				h.Logger.Error("Failed to create user", "user", email, "domain", "user", "target_user", req.Email, "error", err)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Password too short",
				}
			case errors.Is(err, shared.ErrEmailContainsPassword):
				h.Logger.Error("Failed to create user", "user", email, "domain", "user", "target_user", req.Email, "error", err)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Email contains password",
				}
			case errors.Is(err, shared.ErrPasswordContainsEmail):
				h.Logger.Error("Failed to create user", "user", email, "domain", "user", "target_user", req.Email, "error", err)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Password contains email",
				}
			default:
				h.Logger.Error("Failed to create user", "user", email, "domain", "user", "target_user", req.Email, "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   fmt.Sprintf("Failed to create user %s", err),
				}
			}

			json.NewEncoder(w).Encode(resp)
			return
		}

		h.Logger.Info("New user created", "user", email, "domain", "user", "target_user", req.Email)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}

func (h *Handler) DeleteUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, email := ctxutil.GetActor(ctx)

		var req DeleteUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.Logger.Error("Malformed request", "user", email, "domain", "user", "target_user", req.UserID, "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Malformed request"})
			return
		}

		email, err := h.Service.GetEmailFromUserID(req.UserID)
		if err != nil {
			h.Logger.Error("Failed to get email from userID", "user", email, "domain", "user", "target_user", req.UserID, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to delete user"})
		}

		var resp shared.ApiResponse[any]
		if err := h.Service.deleteUser(req.UserID, ctx); err != nil {
			if errors.Is(err, shared.ErrCannotDeleteOwnUser) {
				h.Logger.Error("Failed to delete user", "user", email, "domain", "user", "target_user", email, "error", err)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Failed to delete user",
				}
			} else {
				h.Logger.Error("Failed to delete user", "user", email, "domain", "user", "target_user", email, "error", err)
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
			h.Logger.Warn("First time user creation blocked", "domain", "user")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "First time user creation blocked"})
			return
		}

		var req CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.Logger.Error("Malformed request", "domain", "user", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Malformed request"})
			return
		}

		req.IsActive = true
		req.IsAdmin = true
		req.APIEnabled = true
		req.UIEnabled = true

		var resp shared.ApiResponse[any]
		if err := h.Service.createUser(req, ctx); err != nil {
			switch {
			case errors.Is(err, shared.ErrEmailPattern):
				h.Logger.Error("Failed to create user", "domain", "user", "target_user", req.Email, "error", err)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Invalid email",
				}
			case errors.Is(err, shared.ErrPasswordLength):
				h.Logger.Error("Failed to create user", "domain", "user", "target_user", req.Email, "error", err)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Password too short",
				}
			case errors.Is(err, shared.ErrEmailContainsPassword):
				h.Logger.Error("Failed to create user", "domain", "user", "target_user", req.Email, "error", err)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Email contains password",
				}
			case errors.Is(err, shared.ErrPasswordContainsEmail):
				h.Logger.Error("Failed to create user", "domain", "user", "target_user", req.Email, "error", err)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Password contains email",
				}
			default:
				h.Logger.Error("Failed to create user", "domain", "user", "target_user", req.Email, "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   fmt.Sprintf("Failed to create user %s", err),
				}
			}
			json.NewEncoder(w).Encode(resp)
			return
		}

		// Disable setup mode, mode will set false automatically on next restart
		h.Service.Config.SetupMode = false
		h.Logger.Info("First time user created", "domain", "user", "target_user", req.Email)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}

func (h *Handler) ChangePassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, email := ctxutil.GetActor(ctx)

		var req ChangePasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.Logger.Error("Malformed request", "user", email, "domain", "user", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Malformed request"})
			return
		}

		var resp shared.ApiResponse[any]
		err := h.Service.changePassword(req, ctx)
		if err != nil {
			switch {
			case errors.Is(err, shared.ErrCurrentOrNewPasswordMissing):
				h.Logger.Error("Failed to change password", "user", email, "domain", "user", "error", err)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Current or new password missing",
				}
			case errors.Is(err, shared.ErrAuthenticationFailed):
				h.Logger.Error("Failed to change password", "user", email, "domain", "user", "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Authentication failed",
				}
			case errors.Is(err, shared.ErrInvalidPassword):
				h.Logger.Error("Failed to change password", "user", email, "domain", "user", "error", err)
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Password did not match complexity requirements",
				}
			case errors.Is(err, shared.ErrNoRowsUpdated):
				h.Logger.Error("Failed to change password", "user", email, "domain", "user", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Password was not changed",
				}
			case errors.Is(err, shared.ErrNoRowsDeleted):
				h.Logger.Warn("Failed to clear sessions after password change", "user", email, "domain", "user", "error", err)
				// Not an actual error
				resp = shared.ApiResponse[any]{
					Success: true,
				}
			default:
				h.Logger.Error("Failed to change password", "user", email, "domain", "user", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Failed to change password",
				}
			}

			json.NewEncoder(w).Encode(resp)
			return
		}

		h.Logger.Info("Password changed", "user", email, "domain", "user")
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}

// Initial MFA enrolment - user served with QR code
func (h *Handler) EnrolMFA() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, email := ctxutil.GetActor(ctx)

		// Get QR code
		var resp shared.ApiResponse[any]
		bytes, err := h.Service.enrolMFA(userID, email)
		if err != nil {
			switch {
			case errors.Is(err, shared.ErrMFANotSupported):
				h.Logger.Warn("MFA enrolment not supported for this user type", "user", email, "domain", "user")
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "MFA enrolment not supported for this user type",
				}
			default:
				h.Logger.Error("Failed to enrol MFA", "user", email, "domain", "user", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Failed to enrol MFA",
				}
			}

			json.NewEncoder(w).Encode(resp)
			return
		}

		h.Logger.Info("MFA enrolment begun", "user", email, "domain", "user")
		w.Header().Set("Content-Type", "image/png")
		w.Write(bytes) // test
		encodedBytes := base64.StdEncoding.EncodeToString(bytes)
		json.NewEncoder(w).Encode(shared.ApiResponse[string]{Success: true, Data: encodedBytes})
	}
}

// Second step of enrolment - user enters code to complete MFA enrolment
func (h *Handler) ConfirmMFA() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, email := ctxutil.GetActor(ctx)

		var req MFARequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.Logger.Error("Malformed request", "user", email, "domain", "user", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Malformed request"})
			return
		}

		req.UserID = userID

		var resp shared.ApiResponse[any]
		if err := h.Service.confirmMFA(req); err != nil {
			switch {
			case errors.Is(err, shared.ErrIncorrectMFACode):
				h.Logger.Warn("MFA code incorrect", "user", email, "domain", "user")
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "MFA code incorrect",
				}
			case errors.Is(err, shared.ErrMFANotEnrolled):
				h.Logger.Warn("MFA must be enrolled before being confirmed", "user", email, "domain", "user")
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "MFA must be enrolled before being confirmed",
				}
			case errors.Is(err, shared.ErrNoRowsUpdated):
				h.Logger.Warn("MFA already confirmed", "user", email, "domain", "user")
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "MFA already confirmed",
				}
			default:
				h.Logger.Error("Failed to confirm MFA", "user", email, "domain", "user", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Failed to confirm MFA",
				}
			}

			json.NewEncoder(w).Encode(resp)
			return
		}

		h.Logger.Info("MFA enrolment confirmed", "user", email, "domain", "user")
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}

func (h *Handler) DeleteMFA() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID, email := ctxutil.GetActor(ctx)

		var req MFARequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.Logger.Error("Malformed request", "user", email, "domain", "user", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Malformed request"})
			return
		}

		req.UserID = userID

		var resp shared.ApiResponse[any]
		if err := h.Service.deleteMFA(req); err != nil {
			switch {
			case errors.Is(err, shared.ErrIncorrectMFACode):
				h.Logger.Warn("MFA code incorrect", "user", email, "domain", "user")
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "MFA code incorrect",
				}
			case errors.Is(err, shared.ErrMFANotEnrolled):
				h.Logger.Warn("MFA must be enrolled before being deleted", "user", email, "domain", "user")
				w.WriteHeader(http.StatusBadRequest)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "MFA must be enrolled before being deleted",
				}
			case errors.Is(err, shared.ErrNoRowsUpdated):
				h.Logger.Warn("No MFA found to delete", "user", email, "domain", "user")
				w.WriteHeader(http.StatusNotFound)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "No MFA found to delete",
				}
			default:
				h.Logger.Error("Failed to delete MFA", "user", email, "domain", "user", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Failed to delete MFA",
				}
			}

			json.NewEncoder(w).Encode(resp)
			return
		}

		h.Logger.Info("MFA deleted", "user", email, "domain", "user")
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}
