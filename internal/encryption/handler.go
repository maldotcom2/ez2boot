package encryption

import (
	"encoding/json"
	"errors"
	"ez2boot/internal/ctxutil"
	"ez2boot/internal/shared"
	"net/http"
)

func (h *Handler) RotateEncryptionPhrase() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, email := ctxutil.GetActor(ctx)

		var req RotateEncryptionPhraseRequest
		json.NewDecoder(r.Body).Decode(&req)

		if err := h.Service.rotateEncryptionPhrase(req, ctx); err != nil {
			if errors.Is(err, shared.ErrFieldMissing) {
				h.Logger.Error("Required field missing", "user", email, "domain", "notification", "error", err)
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Required field missing"})
				return
			}
			h.Logger.Error("Failed to rotate encryption phrase", "user", email, "domain", "notification", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: false, Error: "Failed to rotate encryption phrase"})
			return
		}

		h.Logger.Info("Rotated encryption phrase", "user", email, "domain", "notification")
		json.NewEncoder(w).Encode(shared.ApiResponse[any]{Success: true})
	}
}
