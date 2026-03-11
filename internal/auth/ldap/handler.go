package ldap

import (
	"encoding/json"
	"errors"
	"ez2boot/internal/ctxutil"
	"ez2boot/internal/shared"
	"net/http"
)

func (h *Handler) GetLdapConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, email := ctxutil.GetActor(ctx)

		c, err := h.Service.getLdapConfig()
		if err != nil {
			if errors.Is(err, shared.ErrLDAPConfigNotFound) {
				h.Logger.Warn("LDAP config not found", "user", email, "domain", "ldap")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "LDAP config not found"})
				return
			}
			h.Logger.Error("Failed to get ldap config", "user", email, "domain", "ldap", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to get ldap config"})
			return
		}

		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: c})
	}
}

func (h *Handler) SetLdapConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, email := ctxutil.GetActor(ctx)

		var req LdapConfigRequest
		json.NewDecoder(r.Body).Decode(&req)

		err := h.Service.setLdapConfig(req)
		if err != nil {
			h.Logger.Error("Failed to set ldap config", "user", email, "domain", "ldap", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to get ldap config"})
			return
		}

		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}

func (h *Handler) DeleteLdapConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, email := ctxutil.GetActor(ctx)

		err := h.Service.deleteLdapConfig()
		if err != nil {
			h.Logger.Error("Failed to delete ldap config", "user", email, "domain", "ldap", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to delete ldap config"})
			return
		}

		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}

func (h *Handler) SearchUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, email := ctxutil.GetActor(ctx)

		var req LdapSearchRequest
		json.NewDecoder(r.Body).Decode(&req)

		user, err := h.Service.SearchUser(req)
		if err != nil && !errors.Is(err, shared.ErrUserNotFound) {
			h.Logger.Error("Failed to search ldap", "user", email, "domain", "ldap", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to search ldap"})
			return
		}

		if errors.Is(err, shared.ErrUserNotFound) {
			h.Logger.Info("Ldap User not found", "user", email, "domain", "ldap")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Ldap User not found"})
			return
		}

		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true, Data: user})
	}
}
