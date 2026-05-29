package galleries

import (
	"encoding/json"
	"errors"
	"net/http"

	galleriesusecase "github.com/Mozlook/fotobudka-backend/internal/galleries"
	"github.com/Mozlook/fotobudka-backend/internal/http/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (h *Handler) PresignGalleryPhotos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "unauthorized",
			"message":    "Brak autoryzacji.",
		})
		return
	}

	galleryID, err := uuid.Parse(r.PathValue("galleryId"))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "invalid_gallery_id",
			"message":    "Niepoprawne ID galerii.",
		})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxGalleryJSONBodyBytes)

	var req presignPhotosRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "invalid_json",
			"message":    "Niepoprawne body JSON.",
		})
		return
	}

	if len(req.Files) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "validation_error",
			"message":    "Lista plików nie może być pusta.",
		})
		return
	}

	if len(req.Files) > 50 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "validation_error",
			"message":    "Za dużo plików w jednym batchu.",
		})
		return
	}

	files := make([]galleriesusecase.PresignGalleryPhotoFile, 0, len(req.Files))

	for _, file := range req.Files {
		if file.Filename == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error_code": "validation_error",
				"message":    "filename jest wymagany.",
			})
			return
		}

		if file.MimeType == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error_code": "validation_error",
				"message":    "mime_type jest wymagany.",
			})
			return
		}

		if file.SizeBytes <= 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error_code": "validation_error",
				"message":    "size_bytes musi być większe od zera.",
			})
			return
		}

		files = append(files, galleriesusecase.PresignGalleryPhotoFile{
			Filename:  file.Filename,
			MimeType:  file.MimeType,
			SizeBytes: file.SizeBytes,
		})
	}

	uploads, err := h.service.PresignGalleryPhotos(
		ctx,
		galleriesusecase.PresignGalleryPhotosInput{
			GalleryID:      galleryID,
			PhotographerID: userID,
			Files:          files,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error_code": "not_found",
				"message":    "Nie znaleziono galerii.",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "presign_failed",
			"message":    "Nie udało się przygotować uploadu zdjęć galerii.",
		})
		return
	}

	responseUploads := make([]presignPhotoUploadResponse, 0, len(uploads))

	for _, upload := range uploads {
		responseUploads = append(responseUploads, presignPhotoUploadResponse{
			PhotoID:   upload.PhotoID,
			PutURL:    upload.PutURL,
			ObjectKey: upload.ObjectKey,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(presignPhotosResponse{
		Uploads: responseUploads,
	})
}

func (h *Handler) CompleteGalleryPhoto(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "unauthorized",
			"message":    "Brak autoryzacji.",
		})
		return
	}

	galleryID, err := uuid.Parse(r.PathValue("galleryId"))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "invalid_gallery_id",
			"message":    "Niepoprawne ID galerii.",
		})
		return
	}

	photoID, err := uuid.Parse(r.PathValue("photoId"))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "invalid_photo_id",
			"message":    "Niepoprawne ID zdjęcia.",
		})
		return
	}

	photo, err := h.service.CompleteGalleryPhotoUpload(
		ctx,
		galleriesusecase.CompleteGalleryPhotoInput{
			GalleryID:      galleryID,
			PhotoID:        photoID,
			PhotographerID: userID,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error_code": "not_found",
				"message":    "Nie znaleziono zdjęcia galerii.",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "complete_failed",
			"message":    "Nie udało się potwierdzić uploadu zdjęcia galerii.",
		})
		return
	}

	imageURL := ""

	if photo.Width > 0 && photo.Height > 0 {
		signedURL, err := h.storage.PresignedGetObject(ctx, photo.ImageKey, presignedDownloadTTL)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error_code": "internal_error",
				"message":    "Nie udało się podpisać zdjęcia galerii.",
			})
			return
		}

		imageURL = signedURL
	}

	response := galleryPhotoResponse{
		ID:        photo.ID,
		ImageURL:  imageURL,
		Width:     photo.Width,
		Height:    photo.Height,
		SortOrder: photo.SortOrder,
		CreatedAt: photo.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func (h *Handler) DeleteGalleryPhoto(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "unauthorized",
			"message":    "Brak autoryzacji.",
		})
		return
	}

	galleryID, err := uuid.Parse(r.PathValue("galleryId"))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "invalid_gallery_id",
			"message":    "Niepoprawne ID galerii.",
		})
		return
	}

	photoID, err := uuid.Parse(r.PathValue("photoId"))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "invalid_photo_id",
			"message":    "Niepoprawne ID zdjęcia.",
		})
		return
	}

	if err := h.repo.DeletePhoto(ctx, photoID, galleryID, userID); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "internal_error",
			"message":    "Nie udało się usunąć zdjęcia galerii.",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
