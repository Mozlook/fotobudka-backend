package sessions

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Mozlook/fotobudka-backend/internal/guard"
	"github.com/Mozlook/fotobudka-backend/internal/http/middleware"
	"github.com/Mozlook/fotobudka-backend/internal/repository/sessions"
	"github.com/google/uuid"
)

type InsertSessionRequest struct {
	Title           string `json:"title"`
	ClientEmail     string `json:"client_email"`
	BasePriceCents  int32  `json:"base_price_cents"`
	IncludedCount   int32  `json:"included_count"`
	ExtraPriceCents int32  `json:"extra_price_cents"`
	MinSelectCount  int32  `json:"min_select_count"`
	Currency        string `json:"currency"`
	PaymentMode     string `json:"payment_mode"`
}

type InsertSessionResponse struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

// GetSession ensures that the authenticated photographer has access to the requested session.
//
// This handler currently performs only the ownership check and returns no content
// when access is allowed.
func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
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

	err = guard.EnsureSessionOwner(r.Context(), h.sessions, sessionID, userID)
	if err != nil {
		if errors.Is(err, guard.ErrSessionNotAccessible) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) InsertSession(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var req InsertSessionRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if req.Currency == "" {
		req.Currency = "PLN"
	}
	if req.PaymentMode == "" {
		req.PaymentMode = "manual"
	}

	h.sessions.InsertSession(r.Context(), sessions.InsertSessionInput{
		PhotographerID:  userID,
		Title:           req.Title,
		ClientEmail:     &req.ClientEmail,
		BasePriceCents:  req.BasePriceCents,
		IncludedCount:   req.IncludedCount,
		ExtraPriceCents: req.ExtraPriceCents,
		MinSelectCount:  req.MinSelectCount,
		Currency:        req.Currency,
		PaymentMode:     req.PaymentMode,
	})
}
