package galleries

import (
	"context"
	"encoding/json"

	dbgen "github.com/Mozlook/fotobudka-backend/internal/platform/db/sqlc"
	"github.com/google/uuid"
)

type Repository struct {
	q *dbgen.Queries
}

func NewRepository(q *dbgen.Queries) *Repository {
	return &Repository{
		q: q,
	}
}

func normalizeJSONRaw(value any) ([]byte, error) {
	switch typed := value.(type) {
	case nil:
		return []byte("{}"), nil
	case []byte:
		if len(typed) == 0 {
			return []byte("{}"), nil
		}

		return typed, nil
	case json.RawMessage:
		if len(typed) == 0 {
			return []byte("{}"), nil
		}

		return typed, nil
	default:
		return json.Marshal(typed)
	}
}

func (r *Repository) GetPublicPhotographerByUsername(
	ctx context.Context,
	username string,
) (PublicPhotographerProfile, error) {
	row, err := r.q.GetPublicPhotographerByUsername(ctx, username)
	if err != nil {
		return PublicPhotographerProfile{}, err
	}

	socialLinks, err := normalizeJSONRaw(row.SocialLinks)
	if err != nil {
		return PublicPhotographerProfile{}, err
	}

	return PublicPhotographerProfile{
		UserID:      row.UserID,
		Username:    row.Username,
		DisplayName: row.DisplayName,
		Bio:         row.Bio,
		SocialLinks: socialLinks,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

func (r *Repository) ListPublicGalleriesByPhotographerID(
	ctx context.Context,
	photographerID uuid.UUID,
) ([]Gallery, error) {
	rows, err := r.q.ListPublicGalleriesByPhotographerID(ctx, photographerID)
	if err != nil {
		return nil, err
	}

	out := make([]Gallery, 0, len(rows))

	for _, row := range rows {
		out = append(out, Gallery{
			ID:             row.ID,
			PhotographerID: row.PhotographerID,
			Title:          row.Title,
			Slug:           row.Slug,
			IsPublic:       row.IsPublic,
			CreatedAt:      row.CreatedAt,
			PhotoCount:     row.PhotoCount,
			CoverImageKey:  row.CoverImageKey,
		})
	}

	return out, nil
}

func (r *Repository) GetPublicGalleryByUsernameAndSlug(
	ctx context.Context,
	username string,
	slug string,
) (Gallery, error) {
	row, err := r.q.GetPublicGalleryByUsernameAndSlug(
		ctx,
		dbgen.GetPublicGalleryByUsernameAndSlugParams{
			Username: username,
			Slug:     slug,
		},
	)
	if err != nil {
		return Gallery{}, err
	}

	return Gallery{
		ID:             row.ID,
		PhotographerID: row.PhotographerID,
		Title:          row.Title,
		Slug:           row.Slug,
		IsPublic:       row.IsPublic,
		CreatedAt:      row.CreatedAt,
	}, nil
}

func (r *Repository) ListPublicGalleryPhotos(
	ctx context.Context,
	galleryID uuid.UUID,
) ([]GalleryPhoto, error) {
	rows, err := r.q.ListPublicGalleryPhotos(ctx, galleryID)
	if err != nil {
		return nil, err
	}

	out := make([]GalleryPhoto, 0, len(rows))

	for _, row := range rows {
		out = append(out, GalleryPhoto{
			ID:        row.ID,
			GalleryID: row.GalleryID,
			ImageKey:  row.ImageKey,
			Width:     *row.Width,
			Height:    *row.Height,
			SortOrder: row.SortOrder,
			CreatedAt: row.CreatedAt,
		})
	}

	return out, nil
}

func (r *Repository) ListByOwner(
	ctx context.Context,
	photographerID uuid.UUID,
) ([]Gallery, error) {
	rows, err := r.q.ListGalleriesByOwner(ctx, photographerID)
	if err != nil {
		return nil, err
	}

	out := make([]Gallery, 0, len(rows))

	for _, row := range rows {
		out = append(out, Gallery{
			ID:             row.ID,
			PhotographerID: row.PhotographerID,
			Title:          row.Title,
			Slug:           row.Slug,
			IsPublic:       row.IsPublic,
			CreatedAt:      row.CreatedAt,
			PhotoCount:     row.PhotoCount,
			CoverImageKey:  row.CoverImageKey,
		})
	}

	return out, nil
}

func (r *Repository) GetByIDForOwner(
	ctx context.Context,
	galleryID uuid.UUID,
	photographerID uuid.UUID,
) (Gallery, error) {
	row, err := r.q.GetGalleryByIDForOwner(
		ctx,
		dbgen.GetGalleryByIDForOwnerParams{
			ID:             galleryID,
			PhotographerID: photographerID,
		},
	)
	if err != nil {
		return Gallery{}, err
	}

	return Gallery{
		ID:             row.ID,
		PhotographerID: row.PhotographerID,
		Title:          row.Title,
		Slug:           row.Slug,
		IsPublic:       row.IsPublic,
		CreatedAt:      row.CreatedAt,
		PhotoCount:     row.PhotoCount,
		CoverImageKey:  row.CoverImageKey,
	}, nil
}

func (r *Repository) Create(
	ctx context.Context,
	input CreateGalleryInput,
) (Gallery, error) {
	row, err := r.q.CreateGallery(
		ctx,
		dbgen.CreateGalleryParams{
			ID:             input.ID,
			PhotographerID: input.PhotographerID,
			Title:          input.Title,
			Slug:           input.Slug,
			IsPublic:       input.IsPublic,
		},
	)
	if err != nil {
		return Gallery{}, err
	}

	return Gallery{
		ID:             row.ID,
		PhotographerID: row.PhotographerID,
		Title:          row.Title,
		Slug:           row.Slug,
		IsPublic:       row.IsPublic,
		CreatedAt:      row.CreatedAt,
	}, nil
}

func (r *Repository) Update(
	ctx context.Context,
	input UpdateGalleryInput,
) (Gallery, error) {
	row, err := r.q.UpdateGallery(
		ctx,
		dbgen.UpdateGalleryParams{
			ID:             input.ID,
			PhotographerID: input.PhotographerID,
			Title:          input.Title,
			Slug:           input.Slug,
			IsPublic:       input.IsPublic,
		},
	)
	if err != nil {
		return Gallery{}, err
	}

	return Gallery{
		ID:             row.ID,
		PhotographerID: row.PhotographerID,
		Title:          row.Title,
		Slug:           row.Slug,
		IsPublic:       row.IsPublic,
		CreatedAt:      row.CreatedAt,
	}, nil
}

func (r *Repository) Delete(
	ctx context.Context,
	galleryID uuid.UUID,
	photographerID uuid.UUID,
) error {
	return r.q.DeleteGallery(
		ctx,
		dbgen.DeleteGalleryParams{
			ID:             galleryID,
			PhotographerID: photographerID,
		},
	)
}

func (r *Repository) ListPhotosForOwner(
	ctx context.Context,
	galleryID uuid.UUID,
	photographerID uuid.UUID,
) ([]GalleryPhoto, error) {
	rows, err := r.q.ListGalleryPhotosForOwner(
		ctx,
		dbgen.ListGalleryPhotosForOwnerParams{
			GalleryID:      galleryID,
			PhotographerID: photographerID,
		},
	)
	if err != nil {
		return nil, err
	}

	out := make([]GalleryPhoto, 0, len(rows))

	for _, row := range rows {
		out = append(out, GalleryPhoto{
			ID:        row.ID,
			GalleryID: row.GalleryID,
			ImageKey:  row.ImageKey,
			Width:     *row.Width,
			Height:    *row.Height,
			SortOrder: row.SortOrder,
			CreatedAt: row.CreatedAt,
		})
	}

	return out, nil
}

func (r *Repository) GetPhotoForOwner(
	ctx context.Context,
	photoID uuid.UUID,
	galleryID uuid.UUID,
	photographerID uuid.UUID,
) (GalleryPhoto, error) {
	row, err := r.q.GetGalleryPhotoForOwner(
		ctx,
		dbgen.GetGalleryPhotoForOwnerParams{
			ID:             photoID,
			GalleryID:      galleryID,
			PhotographerID: photographerID,
		},
	)
	if err != nil {
		return GalleryPhoto{}, err
	}

	return GalleryPhoto{
		ID:        row.ID,
		GalleryID: row.GalleryID,
		ImageKey:  row.ImageKey,
		Width:     *row.Width,
		Height:    *row.Height,
		SortOrder: row.SortOrder,
		CreatedAt: row.CreatedAt,
	}, nil
}

func (r *Repository) CreatePhotoFromCompletedUpload(
	ctx context.Context,
	input CreateGalleryPhotoInput,
) (GalleryPhoto, error) {
	row, err := r.q.CreateGalleryPhotoFromCompletedUpload(
		ctx,
		dbgen.CreateGalleryPhotoFromCompletedUploadParams{
			ID:        input.ID,
			GalleryID: input.GalleryID,
			ImageKey:  input.ImageKey,
			Width:     &input.Width,
			Height:    &input.Height,
		},
	)
	if err != nil {
		return GalleryPhoto{}, err
	}

	return GalleryPhoto{
		ID:        row.ID,
		GalleryID: row.GalleryID,
		ImageKey:  row.ImageKey,
		Width:     *row.Width,
		Height:    *row.Height,
		SortOrder: row.SortOrder,
		CreatedAt: row.CreatedAt,
	}, nil
}

func (r *Repository) DeletePhoto(
	ctx context.Context,
	photoID uuid.UUID,
	galleryID uuid.UUID,
	photographerID uuid.UUID,
) error {
	return r.q.DeleteGalleryPhoto(
		ctx,
		dbgen.DeleteGalleryPhotoParams{
			ID:             photoID,
			GalleryID:      galleryID,
			PhotographerID: photographerID,
		},
	)
}

func (r *Repository) ListFeaturedPublicGalleries(
	ctx context.Context,
	limit int32,
) ([]FeaturedPublicGallery, error) {
	if limit <= 0 {
		limit = 4
	}

	if limit > 12 {
		limit = 12
	}

	rows, err := r.q.ListFeaturedPublicGalleries(ctx, limit)
	if err != nil {
		return nil, err
	}

	out := make([]FeaturedPublicGallery, 0, len(rows))

	for _, row := range rows {
		out = append(out, FeaturedPublicGallery{
			ID:                      row.ID,
			PhotographerID:          row.PhotographerID,
			Title:                   row.Title,
			Slug:                    row.Slug,
			IsPublic:                row.IsPublic,
			CreatedAt:               row.CreatedAt,
			PhotoCount:              row.PhotoCount,
			CoverImageKey:           row.CoverImageKey,
			PhotographerUsername:    row.PhotographerUsername,
			PhotographerDisplayName: row.PhotographerDisplayName,
		})
	}

	return out, nil
}
