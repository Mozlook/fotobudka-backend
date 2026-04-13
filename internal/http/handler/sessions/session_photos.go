package sessions

import (
	"encoding/json"
	"errors"
	"fmt"
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

	sessionIDString := r.PathValue("sessionId")
	if sessionIDString == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	sessionID, err := uuid.Parse(sessionIDString)
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

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var requestBody PhotosPresignRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&requestBody); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if len(requestBody.Files) == 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	uploads, err := h.sessionPhotos.PrepareSessionPhotoUploads(r.Context(), sessionID, requestBody.Files)
	if err != nil {
		fmt.Printf("photos presign error: session_id=%s user_id=%s err=%v\n", sessionID, userID, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response := PhotosPresignResponse{
		Uploads: make([]PhotosPresignUploadResponse, len(uploads)),
	}

	for i, upload := range uploads {
		item := PhotosPresignUploadResponse{
			Error: upload.Error,
		}

		if !upload.Error {
			item.PhotoID = upload.PhotoID.String()
			item.PutURL = upload.PutURL.String()
			item.ObjectKey = upload.ObjectKey
		}

		response.Uploads[i] = item
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
