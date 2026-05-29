package galleries

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Mozlook/fotobudka-backend/internal/http/middleware"
	galleriesrepo "github.com/Mozlook/fotobudka-backend/internal/repository/galleries"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	maxGalleryJSONBodyBytes = 1 << 20
	presignedDownloadTTL    = 30 * time.Minute
)

func (h *Handler) ListGalleries(w http.ResponseWriter, r *http.Request) {
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

	galleries, err := h.repo.ListByOwner(ctx, userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "internal_error",
			"message":    "Nie udało się pobrać galerii.",
		})
		return
	}

	items := make([]galleryResponse, 0, len(galleries))

	for _, gallery := range galleries {
		coverURL := ""

		if gallery.CoverImageKey != "" {
			signedURL, err := h.storage.PresignedGetObject(ctx, gallery.CoverImageKey, presignedDownloadTTL)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error_code": "internal_error",
					"message":    "Nie udało się podpisać okładki galerii.",
				})
				return
			}

			coverURL = signedURL
		}

		items = append(items, galleryResponse{
			ID:         gallery.ID,
			Title:      gallery.Title,
			Slug:       gallery.Slug,
			IsPublic:   gallery.IsPublic,
			PhotoCount: gallery.PhotoCount,
			CoverURL:   coverURL,
			CreatedAt:  gallery.CreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"Galleries": items,
	})
}

func (h *Handler) CreateGallery(w http.ResponseWriter, r *http.Request) {
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

	r.Body = http.MaxBytesReader(w, r.Body, maxGalleryJSONBodyBytes)

	var req upsertGalleryRequest

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

	req.Title = strings.TrimSpace(req.Title)
	req.Slug = normalizeGallerySlug(req.Slug)

	if err := validateGalleryInput(req.Title, req.Slug); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "validation_error",
			"message":    err.Error(),
		})
		return
	}

	gallery, err := h.repo.Create(
		ctx,
		galleriesrepo.CreateGalleryInput{
			ID:             uuid.New(),
			PhotographerID: userID,
			Title:          req.Title,
			Slug:           req.Slug,
			IsPublic:       req.IsPublic,
		},
	)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error_code": "slug_taken",
				"message":    "Galeria o takim slugu już istnieje.",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "internal_error",
			"message":    "Nie udało się utworzyć galerii.",
		})
		return
	}

	response := galleryResponse{
		ID:         gallery.ID,
		Title:      gallery.Title,
		Slug:       gallery.Slug,
		IsPublic:   gallery.IsPublic,
		PhotoCount: gallery.PhotoCount,
		CreatedAt:  gallery.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetGallery(w http.ResponseWriter, r *http.Request) {
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

	gallery, err := h.repo.GetByIDForOwner(ctx, galleryID, userID)
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
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "internal_error",
			"message":    "Nie udało się pobrać galerii.",
		})
		return
	}

	coverURL := ""

	if gallery.CoverImageKey != "" {
		signedURL, err := h.storage.PresignedGetObject(ctx, gallery.CoverImageKey, presignedDownloadTTL)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error_code": "internal_error",
				"message":    "Nie udało się podpisać okładki galerii.",
			})
			return
		}

		coverURL = signedURL
	}

	photos, err := h.repo.ListPhotosForOwner(ctx, galleryID, userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "internal_error",
			"message":    "Nie udało się pobrać zdjęć galerii.",
		})
		return
	}

	photoResponses := make([]galleryPhotoResponse, 0, len(photos))

	for _, photo := range photos {
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

		photoResponses = append(photoResponses, galleryPhotoResponse{
			ID:        photo.ID,
			ImageURL:  imageURL,
			Width:     photo.Width,
			Height:    photo.Height,
			SortOrder: photo.SortOrder,
			CreatedAt: photo.CreatedAt,
		})
	}

	galleryResponse := galleryResponse{
		ID:         gallery.ID,
		Title:      gallery.Title,
		Slug:       gallery.Slug,
		IsPublic:   gallery.IsPublic,
		PhotoCount: gallery.PhotoCount,
		CoverURL:   coverURL,
		CreatedAt:  gallery.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"Gallery": galleryResponse,
		"Photos":  photoResponses,
	})
}

func (h *Handler) UpdateGallery(w http.ResponseWriter, r *http.Request) {
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

	var req upsertGalleryRequest

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

	req.Title = strings.TrimSpace(req.Title)
	req.Slug = normalizeGallerySlug(req.Slug)

	if err := validateGalleryInput(req.Title, req.Slug); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "validation_error",
			"message":    err.Error(),
		})
		return
	}

	gallery, err := h.repo.Update(
		ctx,
		galleriesrepo.UpdateGalleryInput{
			ID:             galleryID,
			PhotographerID: userID,
			Title:          req.Title,
			Slug:           req.Slug,
			IsPublic:       req.IsPublic,
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

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error_code": "slug_taken",
				"message":    "Galeria o takim slugu już istnieje.",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "internal_error",
			"message":    "Nie udało się zaktualizować galerii.",
		})
		return
	}

	response := galleryResponse{
		ID:         gallery.ID,
		Title:      gallery.Title,
		Slug:       gallery.Slug,
		IsPublic:   gallery.IsPublic,
		PhotoCount: gallery.PhotoCount,
		CreatedAt:  gallery.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func (h *Handler) DeleteGallery(w http.ResponseWriter, r *http.Request) {
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

	if err := h.repo.Delete(ctx, galleryID, userID); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "internal_error",
			"message":    "Nie udało się usunąć galerii.",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
