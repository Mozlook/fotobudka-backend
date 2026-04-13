package client

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/Mozlook/fotobudka-backend/internal/http/middleware"
	"github.com/Mozlook/fotobudka-backend/internal/platform/captcha"
	"github.com/Mozlook/fotobudka-backend/internal/sessionaccess"
	"github.com/Mozlook/fotobudka-backend/internal/sessionphotos"
	"github.com/google/uuid"
)

type GetSessionPhotosResponse struct {
	Items  []sessionphotos.ClientSessionPhotoResponse `json:"items"`
	Offset int32                                      `json:"offset"`
	Limit  int32                                      `json:"limit"`
}

type ClientSessionByCodeRequest struct {
	Code         string `json:"code"`
	CaptchaToken string `json:"captcha_token"`
}

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

	clientSession, err := h.sessionAccess.GetClientSessionByToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, sessionaccess.ErrSessionAccessNotFound) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	payload, err := json.Marshal(ClientSessionResponse{
		ID:              clientSession.SessionID,
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

	cookieToken, expiresAt, err := h.clientManager.IssueClientToken(clientSession.SessionAccessID, clientSession.SessionID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	h.clientManager.SetClientCookie(w, cookieToken, expiresAt)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}

func (h *Handler) GetSessionByCode(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 8*1024)
	defer r.Body.Close()

	var requestBody ClientSessionByCodeRequest

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&requestBody); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "request body is invalid")
		return
	}

	code := strings.TrimSpace(requestBody.Code)
	if code == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "code is required")
		return
	}

	captchaToken := strings.TrimSpace(requestBody.CaptchaToken)

	ip := r.RemoteAddr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		ip = host
	}

	captchaRequired, err := h.redis.RequiresCodeCaptcha(r.Context(), ip)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}

	if captchaRequired {
		if captchaToken == "" {
			writeAPIError(w, http.StatusBadRequest, "captcha_required", "captcha is required")
			return
		}

		ok, err := captcha.Verify(r.Context(), h.recaptchaSecretKey, captchaToken, ip)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "captcha_verification_failed", "captcha verification failed")
			return
		}

		if !ok {
			writeAPIError(w, http.StatusBadRequest, "invalid_captcha", "captcha token is invalid")
			return
		}
	}

	clientSession, err := h.sessionAccess.GetClientSessionByCode(r.Context(), code)
	if err != nil {
		if errors.Is(err, sessionaccess.ErrSessionAccessNotFound) {
			if _, redisErr := h.redis.RegisterFailedCodeAttempt(r.Context(), ip); redisErr != nil {
				writeAPIError(w, http.StatusInternalServerError, "internal_error", "internal error")
				return
			}

			writeAPIError(w, http.StatusNotFound, "session_access_not_found", "session access was not found")
			return
		}

		writeAPIError(w, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}

	if err := h.redis.ClearFailedCodeAttempts(r.Context(), ip); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "internal error")
		return
	}

	cookieToken, expiresAt, err := h.clientManager.IssueClientToken(clientSession.SessionAccessID, clientSession.SessionID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	h.clientManager.SetClientCookie(w, cookieToken, expiresAt)
	writeJSON(w, http.StatusOK, ClientSessionResponse{
		ID:              clientSession.SessionID,
		Status:          clientSession.Status,
		BasePriceCents:  clientSession.BasePriceCents,
		IncludedCount:   clientSession.IncludedCount,
		ExtraPriceCents: clientSession.ExtraPriceCents,
		MinSelectCount:  clientSession.MinSelectCount,
		Currency:        clientSession.Currency,
		PaymentMode:     clientSession.PaymentMode,
		Title:           clientSession.Title,
	})
}

func (h *Handler) GetSessionPhotos(w http.ResponseWriter, r *http.Request) {
	offset := r.URL.Query().Get("offset")
	offsetCount := 0
	if offset != "" {
		offsetCount, err := strconv.Atoi(offset)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if offsetCount < 0 {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
	}

	sessionID, ok := middleware.ClientSessionIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(photos)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
