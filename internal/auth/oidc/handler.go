package oidc

import (
	"encoding/json"
	"errors"
	"ez2boot/internal/ctxutil"
	"ez2boot/internal/shared"
	"ez2boot/internal/util"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func (h *Handler) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// User context not available yet

		if h.Service.Provider == nil {
			h.Logger.Warn("OIDC login attempted but provider not configured", "domain", "oidc")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "SSO is not configured"})
			return
		}

		// Generate state parameter to prevent CSRF
		state, err := util.GenerateRandomString(32)
		if err != nil {
			h.Logger.Error("Failed to generate state", "domain", "oidc", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to initiate SSO login"})
			return
		}

		// Store state in cookie for verification in callback
		http.SetCookie(w, &http.Cookie{
			Name:     "oidc_state",
			Value:    state,
			MaxAge:   int((5 * time.Minute).Seconds()),
			SameSite: h.Config.SameSiteMode,
			HttpOnly: true,
			Secure:   h.Config.SecureCookie,
		})

		h.Logger.Info("Oidc login initiated", "domain", "oidc")
		http.Redirect(w, r, h.Service.Provider.AuthCodeURL(state), http.StatusFound)
	}
}

func (h *Handler) Callback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Verify state to prevent CSRF
		stateCookie, err := r.Cookie("oidc_state")
		if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
			h.Logger.Warn("OIDC state mismatch", "domain", "oidc")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Invalid state"})
			return
		}

		// Exchange code for tokens
		token, err := h.Service.Provider.Exchange(ctx, r.URL.Query().Get("code"))
		if err != nil {
			h.Logger.Error("Failed to exchange code", "domain", "oidc", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to exchange code"})
			return
		}

		// Verify ID token and extract claims
		claims, err := h.Service.Provider.VerifyIDToken(ctx, token)
		if err != nil {
			h.Logger.Error("Failed to verify ID token", "domain", "oidc", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to verify token"})
			return
		}

		// Extract email from claims
		email, ok := claims["email"].(string)
		if !ok || email == "" {
			email, ok = claims["preferred_username"].(string)
			if !ok || email == "" {
				h.Logger.Error("No email claim in token", "domain", "oidc")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "No email in token"})
				return
			}
		}

		// Provision or login user
		var resp shared.ApiResponse[any]
		sessionToken, err := h.Service.loginOidcUser(email, ctx)
		if err != nil {
			switch {
			case errors.Is(err, shared.ErrUserInactive):
				h.Logger.Warn("Login failed", "user", email, "domain", "oidc", "error", err)
				w.WriteHeader(http.StatusForbidden)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "User not authorised",
				}
			case errors.Is(err, shared.ErrUserNotAuthorised):
				h.Logger.Warn("Login failed", "user", email, "domain", "oidc", "error", err)
				w.WriteHeader(http.StatusForbidden)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "User not authorised",
				}
			case errors.Is(err, shared.ErrWrongIdentityProvider):
				h.Logger.Warn("Login failed", "user", email, "domain", "oidc", "error", err)
				w.WriteHeader(http.StatusForbidden)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "User not authorised",
				}
			default:
				h.Logger.Error("Login failed", "user", email, "domain", "oidc", "error", err)
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
			Value:    sessionToken,
			Path:     "/",
			Expires:  time.Now().Add(h.Config.UserSessionDuration),
			SameSite: h.Config.SameSiteMode,
			HttpOnly: true,
			Secure:   h.Config.SecureCookie,
		})

		h.Logger.Info("User logged in", "user", email, "domain", "oidc")
		// Redirect to app
		if strings.Contains(h.Version, "dev") { // For local dev with Vite
			http.Redirect(w, r, "http://localhost:5173/dashboard", http.StatusFound)
		} else {
			http.Redirect(w, r, "/dashboard", http.StatusFound)
		}
	}
}

func (h *Handler) TestOidcConnection() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, email := ctxutil.GetActor(ctx)

		var resp shared.ApiResponse[any]
		if err := h.Service.testOidcConnection(ctx); err != nil {
			switch {
			case errors.Is(err, shared.ErrOIDCConfigNotFound):
				h.Logger.Warn("OIDC test failed", "user", email, "domain", "oidc", "error", err)
				w.WriteHeader(http.StatusNotFound)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "OIDC is not configured",
				}
			default:
				h.Logger.Error("OIDC test failed", "user", email, "domain", "oidc", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   fmt.Sprintf("OIDC connection test failed %s", err),
				}
			}

			json.NewEncoder(w).Encode(resp)
			return
		}

		h.Logger.Info("OIDC connection test succeeded", "user", email, "domain", "oidc")
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}

func (h *Handler) GetOidcConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, email := ctxutil.GetActor(ctx)

		var resp shared.ApiResponse[any]
		c, err := h.Service.getOidcConfig()
		if err != nil {
			switch {
			case errors.Is(err, shared.ErrOIDCConfigNotFound):
				h.Logger.Warn("Oidc config not found", "user", email, "domain", "oidc", "error", err)
				w.WriteHeader(http.StatusOK)
				resp = shared.ApiResponse[any]{
					Success: true, // No config is not an error
					Data:    nil,
				}
			default:
				h.Logger.Error("Failed to get oidc config", "user", email, "domain", "oidc", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Failed to get oidc config",
				}
			}

			json.NewEncoder(w).Encode(resp)
			return
		}

		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: c})
	}
}

func (h *Handler) SetOidcConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, email := ctxutil.GetActor(ctx)

		var req OidcConfigRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.Logger.Error("Malformed request", "user", email, "domain", "oidc", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Malformed request"})
			return
		}

		err := h.Service.setOidcConfig(req, ctx)
		if err != nil {
			h.Logger.Error("Failed to set oidc config", "user", email, "domain", "oidc", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to get oidc config"})
			return
		}

		h.Logger.Info("Oidc config set", "user", email, "domain", "oidc")
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}

func (h *Handler) DeleteOidcConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, email := ctxutil.GetActor(ctx)

		// There's only one config to delete, no payload as selector

		if err := h.Service.deleteOidcConfig(ctx); err != nil {
			h.Logger.Error("Failed to delete oidc config", "user", email, "domain", "oidc", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to delete oidc config"})
			return
		}

		h.Logger.Info("Oidc config deleted", "user", email, "domain", "oidc")
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}

func (h *Handler) HasOidc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hasOidc := h.Service.hasOidc()

		response := HasOidcRespose{HasOidc: hasOidc}
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: response})
	}
}
