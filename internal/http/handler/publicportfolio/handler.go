package publicportfolio

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Mozlook/fotobudka-backend/internal/platform/storage"
	galleriesrepo "github.com/Mozlook/fotobudka-backend/internal/repository/galleries"
	"github.com/jackc/pgx/v5"
)

type Handler struct {
	galleriesRepo *galleriesrepo.Repository
	storage       *storage.Client
}

func New(
	galleriesRepo *galleriesrepo.Repository,
	storageClient *storage.Client,
) *Handler {
	return &Handler{
		galleriesRepo: galleriesRepo,
		storage:       storageClient,
	}
}

const presignedDownloadTTL = 30 * time.Minute

type publicProfileResponse struct {
	Username    string          `json:"Username"`
	DisplayName string          `json:"DisplayName"`
	Bio         string          `json:"Bio"`
	SocialLinks json.RawMessage `json:"SocialLinks"`
}

type publicGalleryListItemResponse struct {
	ID         string `json:"ID"`
	Title      string `json:"Title"`
	Slug       string `json:"Slug"`
	PhotoCount int32  `json:"PhotoCount"`
	CoverURL   string `json:"CoverURL,omitempty"`
}

type publicGalleryResponse struct {
	ID    string `json:"ID"`
	Title string `json:"Title"`
	Slug  string `json:"Slug"`
}

type publicGalleryPhotoResponse struct {
	ID        string `json:"ID"`
	ImageURL  string `json:"ImageURL"`
	Width     int32  `json:"Width"`
	Height    int32  `json:"Height"`
	SortOrder int32  `json:"SortOrder"`
}

type featuredGalleryPhotographerResponse struct {
	Username    string `json:"Username"`
	DisplayName string `json:"DisplayName"`
}

type featuredGalleryResponse struct {
	ID           string                              `json:"ID"`
	Title        string                              `json:"Title"`
	Slug         string                              `json:"Slug"`
	PhotoCount   int32                               `json:"PhotoCount"`
	CoverURL     string                              `json:"CoverURL"`
	Photographer featuredGalleryPhotographerResponse `json:"Photographer"`
}

func (h *Handler) GetPhotographer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	username := r.PathValue("username")
	if username == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "invalid_username",
			"message":    "Username jest wymagany.",
		})
		return
	}

	profile, err := h.galleriesRepo.GetPublicPhotographerByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error_code": "not_found",
				"message":    "Nie znaleziono fotografa.",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "internal_error",
			"message":    "Nie udało się pobrać profilu fotografa.",
		})
		return
	}

	galleries, err := h.galleriesRepo.ListPublicGalleriesByPhotographerID(ctx, profile.UserID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "internal_error",
			"message":    "Nie udało się pobrać galerii fotografa.",
		})
		return
	}

	galleryResponses := make([]publicGalleryListItemResponse, 0, len(galleries))

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

		galleryResponses = append(galleryResponses, publicGalleryListItemResponse{
			ID:         gallery.ID.String(),
			Title:      gallery.Title,
			Slug:       gallery.Slug,
			PhotoCount: gallery.PhotoCount,
			CoverURL:   coverURL,
		})
	}

	socialLinks := json.RawMessage(profile.SocialLinks)
	if len(socialLinks) == 0 {
		socialLinks = json.RawMessage("{}")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"Profile": publicProfileResponse{
			Username:    profile.Username,
			DisplayName: profile.DisplayName,
			Bio:         profile.Bio,
			SocialLinks: socialLinks,
		},
		"Galleries": galleryResponses,
	})
}

func (h *Handler) GetGallery(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	username := r.PathValue("username")
	if username == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "invalid_username",
			"message":    "Username jest wymagany.",
		})
		return
	}

	slug := r.PathValue("slug")
	if slug == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "invalid_slug",
			"message":    "Slug galerii jest wymagany.",
		})
		return
	}

	profile, err := h.galleriesRepo.GetPublicPhotographerByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error_code": "not_found",
				"message":    "Nie znaleziono fotografa.",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "internal_error",
			"message":    "Nie udało się pobrać profilu fotografa.",
		})
		return
	}

	gallery, err := h.galleriesRepo.GetPublicGalleryByUsernameAndSlug(ctx, username, slug)
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

	photos, err := h.galleriesRepo.ListPublicGalleryPhotos(ctx, gallery.ID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "internal_error",
			"message":    "Nie udało się pobrać zdjęć galerii.",
		})
		return
	}

	photoResponses := make([]publicGalleryPhotoResponse, 0, len(photos))

	for _, photo := range photos {
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

		photoResponses = append(photoResponses, publicGalleryPhotoResponse{
			ID:        photo.ID.String(),
			ImageURL:  signedURL,
			Width:     photo.Width,
			Height:    photo.Height,
			SortOrder: photo.SortOrder,
		})
	}

	socialLinks := json.RawMessage(profile.SocialLinks)
	if len(socialLinks) == 0 {
		socialLinks = json.RawMessage("{}")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"Profile": publicProfileResponse{
			Username:    profile.Username,
			DisplayName: profile.DisplayName,
			Bio:         profile.Bio,
			SocialLinks: socialLinks,
		},
		"Gallery": publicGalleryResponse{
			ID:    gallery.ID.String(),
			Title: gallery.Title,
			Slug:  gallery.Slug,
		},
		"Photos": photoResponses,
	})
}

func (h *Handler) GetFeaturedGalleries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := int32(4)

	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil || parsedLimit < 1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error_code": "invalid_limit",
				"message":    "Limit musi być liczbą większą od zera.",
			})
			return
		}

		if parsedLimit > 12 {
			parsedLimit = 12
		}

		limit = int32(parsedLimit)
	}

	galleries, err := h.galleriesRepo.ListFeaturedPublicGalleries(ctx, limit)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error_code": "internal_error",
			"message":    "Nie udało się pobrać wyróżnionych galerii.",
		})
		return
	}

	galleryResponses := make([]featuredGalleryResponse, 0, len(galleries))

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

		if coverURL == "" {
			continue
		}

		galleryResponses = append(galleryResponses, featuredGalleryResponse{
			ID:         gallery.ID.String(),
			Title:      gallery.Title,
			Slug:       gallery.Slug,
			PhotoCount: gallery.PhotoCount,
			CoverURL:   coverURL,
			Photographer: featuredGalleryPhotographerResponse{
				Username:    gallery.PhotographerUsername,
				DisplayName: gallery.PhotographerDisplayName,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"Galleries": galleryResponses,
	})
}
