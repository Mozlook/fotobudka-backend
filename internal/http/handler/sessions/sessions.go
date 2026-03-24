package sessions

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Mozlook/fotobudka-backend/internal/guard"
	"github.com/Mozlook/fotobudka-backend/internal/http/middleware"
	"github.com/Mozlook/fotobudka-backend/internal/repository/sessions"
	"github.com/google/uuid"
)

// InsertSessionRequest describes the payload used to create a new session.
type InsertSessionRequest struct {
	Title           string  `json:"title"`
	ClientEmail     *string `json:"client_email"`
	BasePriceCents  int32   `json:"base_price_cents"`
	IncludedCount   int32   `json:"included_count"`
	ExtraPriceCents int32   `json:"extra_price_cents"`
	MinSelectCount  int32   `json:"min_select_count"`
	Currency        string  `json:"currency"`
	PaymentMode     string  `json:"payment_mode"`
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

// InsertSession creates a new session for the authenticated photographer.
//
// The photographer identifier is taken from the request context, not from
// the request body. On success, the handler returns the created session ID
// and its initial status.
func (h *Handler) InsertSession(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
	defer r.Body.Close()

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

	var clientEmail *string
	if req.ClientEmail != nil {
		trimmed := strings.TrimSpace(*req.ClientEmail)
		if trimmed != "" {
			clientEmail = &trimmed
		}
	}

	sessionStatus, err := h.sessions.InsertSession(r.Context(), sessions.InsertSessionInput{
		PhotographerID:  userID,
		Title:           req.Title,
		ClientEmail:     clientEmail,
		BasePriceCents:  req.BasePriceCents,
		IncludedCount:   req.IncludedCount,
		ExtraPriceCents: req.ExtraPriceCents,
		MinSelectCount:  req.MinSelectCount,
		Currency:        req.Currency,
		PaymentMode:     req.PaymentMode,
	})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	payload, err := json.Marshal(sessionStatus)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(payload)
}
