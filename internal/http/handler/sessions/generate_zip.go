package sessions

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Mozlook/fotobudka-backend/internal/deliveries"
	"github.com/Mozlook/fotobudka-backend/internal/guard"
	"github.com/Mozlook/fotobudka-backend/internal/http/middleware"
	"github.com/google/uuid"
)

type GenerateZIPResponse struct {
	DeliveryID uuid.UUID `json:"delivery_id"`
	Version    int32     `json:"version"`
	Status     string    `json:"status"`
}

func (h *Handler) GenerateZIP(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	sessionID, err := uuid.Parse(r.PathValue("sessionId"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = guard.EnsureSessionOwner(r.Context(), h.sessionsRepo, sessionID, userID)
	if err != nil {
		if errors.Is(err, guard.ErrSessionNotAccessible) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	result, err := h.deliveries.GenerateZIP(r.Context(), sessionID)
	if err != nil {
		switch {
		case errors.Is(err, deliveries.ErrInvalidSessionID):
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return

		case errors.Is(err, deliveries.ErrSessionNotFound):
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return

		case errors.Is(err, deliveries.ErrGenerateZIPLocked),
			errors.Is(err, deliveries.ErrNoSelections),
			errors.Is(err, deliveries.ErrMissingFinalPhotos):
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return

		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	payload := GenerateZIPResponse{
		DeliveryID: result.DeliveryID,
		Version:    result.Version,
		Status:     result.Status,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		return
	}
}
