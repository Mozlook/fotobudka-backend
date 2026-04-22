package sessions

import (
	"errors"
	"net/http"

	"github.com/Mozlook/fotobudka-backend/internal/finalphotos"
	"github.com/Mozlook/fotobudka-backend/internal/guard"
	"github.com/Mozlook/fotobudka-backend/internal/http/middleware"
	"github.com/google/uuid"
)

func (h *Handler) CompleteFinalPhotoUpload(w http.ResponseWriter, r *http.Request) {
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

	finalID, err := uuid.Parse(r.PathValue("finalId"))
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

	err = h.finalPhotos.CompleteFinalPhotoUpload(r.Context(), sessionID, finalID)
	if err != nil {
		switch {
		case errors.Is(err, finalphotos.ErrInvalidSessionID),
			errors.Is(err, finalphotos.ErrInvalidFinalID):
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return

		case errors.Is(err, finalphotos.ErrSessionNotFound),
			errors.Is(err, finalphotos.ErrFinalPhotoNotFound):
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return

		case errors.Is(err, finalphotos.ErrFinalUploadLocked),
			errors.Is(err, finalphotos.ErrUploadedObjectNotFound):
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return

		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
