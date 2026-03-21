package oidc

import (
	"encoding/json"
	"errors"
	"ez2boot/internal/ctxutil"
	"ez2boot/internal/shared"
	"net/http"
)

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

		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}
