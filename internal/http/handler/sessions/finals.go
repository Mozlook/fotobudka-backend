package sessions

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Mozlook/fotobudka-backend/internal/finalphotos"
	"github.com/Mozlook/fotobudka-backend/internal/guard"
	"github.com/Mozlook/fotobudka-backend/internal/http/middleware"
	"github.com/google/uuid"
)

type presignFinalFileRequest struct {
	PhotoID   uuid.UUID `json:"photo_id"`
	Filename  string    `json:"filename"`
	MimeType  string    `json:"mime_type"`
	SizeBytes *int64    `json:"size_bytes"`
}

type presignFinalsRequest struct {
	Files []presignFinalFileRequest `json:"files"`
}

type presignFinalUploadResponse struct {
	FinalID   uuid.UUID `json:"final_id"`
	PhotoID   uuid.UUID `json:"photo_id"`
	PutURL    string    `json:"put_url"`
	ObjectKey string    `json:"object_key"`
}

type presignFinalsResponse struct {
	Uploads []presignFinalUploadResponse `json:"uploads"`
}

func (h *Handler) PresignFinals(w http.ResponseWriter, r *http.Request) {
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

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB na JSON metadata
	defer r.Body.Close()

	var req presignFinalsRequest

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if len(req.Files) == 0 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	files := make([]finalphotos.FileInput, 0, len(req.Files))
	for _, file := range req.Files {
		files = append(files, finalphotos.FileInput{
			PhotoID:   file.PhotoID,
			Filename:  strings.TrimSpace(file.Filename),
			MimeType:  strings.TrimSpace(file.MimeType),
			SizeBytes: file.SizeBytes,
		})
	}

	uploads, err := h.finalPhotos.PrepareFinalPhotoUploads(r.Context(), sessionID, files)
	if err != nil {
		switch {
		case errors.Is(err, finalphotos.ErrInvalidSessionID),
			errors.Is(err, finalphotos.ErrEmptyFiles),
			errors.Is(err, finalphotos.ErrInvalidPhotoID),
			errors.Is(err, finalphotos.ErrDuplicatePhotoInBatch),
			errors.Is(err, finalphotos.ErrInvalidFilename),
			errors.Is(err, finalphotos.ErrInvalidMimeType),
			errors.Is(err, finalphotos.ErrInvalidFileSize):
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return

		case errors.Is(err, finalphotos.ErrSessionNotFound),
			errors.Is(err, finalphotos.ErrPhotoNotSelected):
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return

		case errors.Is(err, finalphotos.ErrFinalUploadLocked),
			errors.Is(err, finalphotos.ErrFinalAlreadyExists):
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return

		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	resp := presignFinalsResponse{
		Uploads: make([]presignFinalUploadResponse, 0, len(uploads)),
	}

	for _, upload := range uploads {
		resp.Uploads = append(resp.Uploads, presignFinalUploadResponse{
			FinalID:   upload.FinalID,
			PhotoID:   upload.PhotoID,
			PutURL:    upload.PutURL,
			ObjectKey: upload.ObjectKey,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		return
	}
}
