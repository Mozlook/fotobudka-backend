package sessions

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Mozlook/fotobudka-backend/internal/guard"
	"github.com/Mozlook/fotobudka-backend/internal/http/middleware"
	"github.com/Mozlook/fotobudka-backend/internal/sessionphotos"
	"github.com/google/uuid"
)

type PhotosPresignRequest struct {
	Files []sessionphotos.FileInput `json:"files"`
}

type PhotosPresignUploadResponse struct {
	PhotoID   string `json:"photo_id,omitempty"`
	PutURL    string `json:"put_url,omitempty"`
	ObjectKey string `json:"object_key,omitempty"`
	Error     bool   `json:"error"`
}

type PhotosPresignResponse struct {
	Uploads []PhotosPresignUploadResponse `json:"uploads"`
}

func (h *Handler) PhotosPresign(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	sessionIDString := r.URL.Query().Get("sessionId")
	if sessionIDString == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	sessionID := uuid.MustParse(sessionIDString)

	err := guard.EnsureSessionOwner(r.Context(), h.sessions, sessionID, userID)
	if err != nil {
		if errors.Is(err, guard.ErrSessionNotAccessible) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return

	}

	var requestBody PhotosPresignRequest

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&requestBody); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	files := requestBody.Files

	urls, err := h.sessionPhotos.PrepareSessionPhotoUploads(r.Context(), sessionID, files)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response := make([]PhotosPresignResponse, 0, len(urls))

	for _, upload := range urls {
		if upload.Error == true {
			continue
		}

		obj := PhotosPresignUploadResponse{
			PhotoID:   upload.PhotoID.String(),
			PutURL:    upload.PutURL.String(),
			ObjectKey: upload.ObjectKey,
		}

		response = append(response, obj)
	}

	payload, err := json.Marshal(struct{ response })
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}
