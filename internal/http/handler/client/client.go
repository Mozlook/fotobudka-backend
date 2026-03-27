package client

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Mozlook/fotobudka-backend/internal/sessionaccess"
	"github.com/google/uuid"
)

type ClientSessionResponse struct {
	ID              uuid.UUID `json:"id"`
	Status          string    `json:"status"`
	BasePriceCents  int32     `json:"base_price_cents"`
	IncludedCount   int32     `json:"included_count"`
	ExtraPriceCents int32     `json:"extra_price_cents"`
	MinSelectCount  int32     `json:"min_select_count"`
	Currency        string    `json:"currency"`
	PaymentMode     string    `json:"payment_mode"`
	Title           string    `json:"title"`
}

func (h *Handler) GetSessionByToken(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	clientSession, err := h.sessionAccess.GetClientSessionByTokenHMAC(r.Context(), token)
	if err != nil {
		if errors.Is(err, sessionaccess.ErrSessionAccessNotFound) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	payload, err := json.Marshal(ClientSessionResponse{
		ID:              clientSession.ID,
		Status:          clientSession.Status,
		BasePriceCents:  clientSession.BasePriceCents,
		IncludedCount:   clientSession.IncludedCount,
		ExtraPriceCents: clientSession.ExtraPriceCents,
		MinSelectCount:  clientSession.MinSelectCount,
		Currency:        clientSession.Currency,
		PaymentMode:     clientSession.PaymentMode,
		Title:           clientSession.Title,
	})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}
