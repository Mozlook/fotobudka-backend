package client

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Mozlook/fotobudka-backend/internal/deliveries"
	"github.com/Mozlook/fotobudka-backend/internal/http/middleware"
	"github.com/google/uuid"
)

type GetLatestDeliveryDownloadResponse struct {
	DeliveryID   uuid.UUID  `json:"delivery_id"`
	Version      int32      `json:"version"`
	DownloadURL  string     `json:"download_url"`
	ZipSizeBytes *int64     `json:"zip_size_bytes,omitempty"`
	GeneratedAt  *time.Time `json:"generated_at,omitempty"`
}

func (h *Handler) GetLatestDeliveryDownload(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := middleware.ClientSessionIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	result, err := h.deliveries.GetLatestDeliveryDownloadURL(r.Context(), sessionID)
	if err != nil {
		switch {
		case errors.Is(err, deliveries.ErrLatestReadyDeliveryNotFound):
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		case errors.Is(err, deliveries.ErrInvalidSessionID):
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	payload := GetLatestDeliveryDownloadResponse{
		DeliveryID:   result.DeliveryID,
		Version:      result.Version,
		DownloadURL:  result.DownloadURL,
		ZipSizeBytes: result.ZipSizeBytes,
		GeneratedAt:  result.GeneratedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		return
	}
}
