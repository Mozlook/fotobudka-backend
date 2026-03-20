package me

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/Mozlook/fotobudka-backend/internal/http/middleware"
	"github.com/Mozlook/fotobudka-backend/internal/repository/profiles"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
)

var usernameRE = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{1,30}[a-z0-9])?$`)

var reservedUsernames = map[string]struct{}{
	"api":    {},
	"auth":   {},
	"public": {},
	"me":     {},
	"login":  {},
	"logout": {},
	"s":      {},
}

// PutProfileRequest describes the JSON payload accepted by PutProfile
type PutProfileRequest struct {
	Username    string               `json:"username"`
	DisplayName string               `json:"display_name"`
	Bio         string               `json:"bio"`
	SocialLinks profiles.SocialLinks `json:"social_links"`
}

// GetProfile returns the authenticated photographer profile as JSON.
//
// The user identity is read from the request context and must be set by
// authentication middleware before this handler is executed.
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	profile, err := h.profiles.GetPhotographerProfileByUserID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	payload, err := json.Marshal(profile)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}

// PutProfile creates or updates the authenticated photographer profile.
//
// The user identity is read from the request context and must be set by
// authentication middleware before this handler is executed.
//
// The request body must contain a valid profile payload. The handler validates
// the username format, reserved usernames, display name length, and bio length
// before persisting the profile.
func (h *Handler) PutProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
	defer r.Body.Close()

	var req PutProfileRequest

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	username := strings.TrimSpace(req.Username)
	displayName := strings.TrimSpace(req.DisplayName)
	bio := strings.TrimSpace(req.Bio)

	if !usernameRE.MatchString(username) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if _, reserved := reservedUsernames[strings.ToLower(username)]; reserved {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if n := utf8.RuneCountInString(displayName); n < 1 || n > 80 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if utf8.RuneCountInString(bio) > 1000 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	profile, err := h.profiles.UpsertPhotographerProfile(r.Context(), profiles.UpsertInput{
		UserID:      userID,
		Username:    username,
		DisplayName: displayName,
		Bio:         bio,
		SocialLinks: req.SocialLinks,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	payload, err := json.Marshal(profile)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}
