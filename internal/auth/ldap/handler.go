package ldap

import (
	"encoding/json"
	"errors"
	"ez2boot/internal/ctxutil"
	"ez2boot/internal/shared"
	"net/http"
)

// @Summary 		Get LDAP config
// @Description 	Allow an administrator to get the current LDAP config
// @Tags 			ldap, admin
// @Success 		200 {object} LdapConfigResponse
// @Failure      	400 {object} string
// @Router       	/auth/ldap [get]
// @Security 		BasicAuth
func (h *Handler) GetLdapConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, email := ctxutil.GetActor(ctx)

		var resp shared.ApiResponse[any]
		c, err := h.Service.getLdapConfig()
		if err != nil {
			switch {
			case errors.Is(err, shared.ErrLDAPConfigNotFound):
				h.Logger.Warn("Ldap config not found", "user", email, "domain", "ldap", "error", err)
				w.WriteHeader(http.StatusOK)
				resp = shared.ApiResponse[any]{
					Success: true, // No config is not an error
					Data:    nil,
				}
			default:
				h.Logger.Error("Failed to get ldap config", "user", email, "domain", "ldap", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Failed to get ldap config",
				}
			}

			json.NewEncoder(w).Encode(resp)
			return
		}

		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: c})
	}
}

// @Summary 		Set LDAP config
// @Description 	Allow an administrator to set an LDAP config
// @Tags 			ldap, admin
// @Accept 			json
// @Param			request body LdapConfigRequest true "Request body"
// @Success 		200 {object} bool
// @Failure      	400 {object} string
// @Router       	/auth/ldap [post]
// @Security 		BasicAuth
func (h *Handler) SetLdapConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, email := ctxutil.GetActor(ctx)

		var req LdapConfigRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.Logger.Error("Malformed request", "user", email, "domain", "ldap", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Malformed request"})
			return
		}

		err := h.Service.setLdapConfig(req, ctx)
		if err != nil {
			h.Logger.Error("Failed to set ldap config", "user", email, "domain", "ldap", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to get ldap config"})
			return
		}

		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}

// @Summary 		Delete LDAP config
// @Description 	Allow an administrator to delete an LDAP config. No payload or selector, as there can be only one configuration.
// @Tags 			ldap, admin
// @Success 		200 {object} bool
// @Router       	/auth/ldap [delete]
// @Security 		BasicAuth
func (h *Handler) DeleteLdapConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, email := ctxutil.GetActor(ctx)

		// There's only one config to delete, no payload as selector

		if err := h.Service.deleteLdapConfig(ctx); err != nil {
			h.Logger.Error("Failed to delete ldap config", "user", email, "domain", "ldap", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to delete ldap config"})
			return
		}

		h.Logger.Info("Ldap config deleted", "user", email, "domain", "ldap")
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}

// @Summary 		Search for LDAP user
// @Description 	Allow an administrator to search the LDAP service for a user during provisioning. Matches where UPN starts with query and returns one match only.
// @Tags 			ldap, admin
// @Accept 			json
// @Param			request body LdapSearchRequest true "Request body"
// @Success 		200 {object} LdapSearchResponse
// @Failure      	400 {object} string
// @Failure      	404 {object} string
// @Failure      	503 {object} string
// @Router       	/auth/ldap/users/search [post]
// @Security 		BasicAuth
func (h *Handler) SearchUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, email := ctxutil.GetActor(ctx)

		var req LdapSearchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.Logger.Error("Malformed request", "user", email, "domain", "ldap", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Malformed request"})
			return
		}

		var resp shared.ApiResponse[any]
		user, err := h.Service.Searcher.SearchUser(req)
		if err != nil {
			switch {
			case errors.Is(err, shared.ErrUserNotFound):
				h.Logger.Info("Ldap User not found", "user", email, "domain", "ldap")
				w.WriteHeader(http.StatusNotFound)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Ldap User not found",
				}
			case errors.Is(err, shared.ErrLDAPConfigNotFound):
				h.Logger.Warn("Search failed", "user", email, "domain", "ldap", "error", err)
				w.WriteHeader(http.StatusServiceUnavailable)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Ldap is not configured",
				}
			default:
				h.Logger.Error("Failed to search ldap", "user", email, "domain", "ldap", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Failed to search ldap",
				}
			}

			json.NewEncoder(w).Encode(resp)
			return
		}

		h.Logger.Info("Ldap search", "user", email, "domain", "ldap", "query", req.Query)
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: user})
	}
}

// @Summary 		Create LDAP user
// @Description 	Allow an administrator to create an LDAP user. Assumes an LDAP config is present.
// @Tags 			ldap, admin
// @Accept 			json
// @Param			request body CreateLdapUserRequest true "Request body"
// @Success 		201 {object} bool
// @Failure      	400 {object} string
// @Failure      	404 {object} string
// @Failure      	409 {object} string
// @Router       	/user/ldap [post]
// @Security 		BasicAuth
func (h *Handler) CreateLdapUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, email := ctxutil.GetActor(ctx)

		var req CreateLdapUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.Logger.Error("Malformed request", "user", email, "domain", "ldap", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Malformed request"})
			return
		}

		var resp shared.ApiResponse[any]
		if err := h.Service.createLdapUser(req.Email, ctx); err != nil {
			switch {
			case errors.Is(err, shared.ErrUserAlreadyExists):
				h.Logger.Error("Failed to create user", "user", email, "domain", "ldap", "target_user", req.Email, "error", err)
				w.WriteHeader(http.StatusConflict)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "User already exists",
				}
			case errors.Is(err, shared.ErrUserNotFound):
				h.Logger.Warn("Failed to create user", "user", email, "domain", "ldap", "target_user", req.Email, "error", err)
				w.WriteHeader(http.StatusNotFound)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "User not found in directory",
				}
			default:
				h.Logger.Error("Failed to create user", "user", email, "domain", "ldap", "target_user", req.Email, "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				resp = shared.ApiResponse[any]{
					Success: false,
					Error:   "Failed to create user",
				}
			}

			json.NewEncoder(w).Encode(resp)
			return
		}

		h.Logger.Info("New user created", "user", email, "domain", "ldap", "target_user", req.Email)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}
