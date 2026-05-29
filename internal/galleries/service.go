package galleries

import (
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"time"

	"github.com/Mozlook/fotobudka-backend/internal/platform/storage"
	galleriesrepo "github.com/Mozlook/fotobudka-backend/internal/repository/galleries"
	sessionphotos "github.com/Mozlook/fotobudka-backend/internal/sessionphotos"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxGalleryPresignBatchSize = 50
	presignedPutTTL            = 30 * time.Minute
)

type Service struct {
	pool    *pgxpool.Pool
	repo    *galleriesrepo.Repository
	storage *storage.Client
}

func New(
	pool *pgxpool.Pool,
	repo *galleriesrepo.Repository,
	storageClient *storage.Client,
) *Service {
	return &Service{
		pool:    pool,
		repo:    repo,
		storage: storageClient,
	}
}

func (s *Service) PresignGalleryPhotos(
	ctx context.Context,
	input PresignGalleryPhotosInput,
) ([]GalleryPhotoUploadTarget, error) {
	if input.GalleryID == uuid.Nil {
		return nil, fmt.Errorf("gallery id is required")
	}

	if input.PhotographerID == uuid.Nil {
		return nil, fmt.Errorf("photographer id is required")
	}

	if len(input.Files) == 0 {
		return nil, fmt.Errorf("files cannot be empty")
	}

	if len(input.Files) > maxGalleryPresignBatchSize {
		return nil, fmt.Errorf("too many files in one batch")
	}

	if _, err := s.repo.GetByIDForOwner(
		ctx,
		input.GalleryID,
		input.PhotographerID,
	); err != nil {
		return nil, err
	}

	uploads := make([]GalleryPhotoUploadTarget, 0, len(input.Files))

	for _, file := range input.Files {
		if file.Filename == "" {
			return nil, fmt.Errorf("filename is required")
		}

		if file.SizeBytes <= 0 {
			return nil, fmt.Errorf("invalid file size for %s", file.Filename)
		}

		ext, ok := sessionphotos.SourceExtFromMIME(file.MimeType)
		if !ok {
			return nil, fmt.Errorf("unsupported mime type for %s", file.Filename)
		}

		photoID := uuid.New()
		objectKey := buildGalleryPhotoObjectKey(input.GalleryID, photoID, ext)

		putURL, err := s.storage.PresignedPutObject(ctx, objectKey, presignedPutTTL)
		if err != nil {
			return nil, err
		}

		if _, err := s.repo.CreatePhotoForPresignedUpload(
			ctx,
			galleriesrepo.CreateGalleryPhotoInput{
				ID:        photoID,
				GalleryID: input.GalleryID,
				ImageKey:  objectKey,
				Width:     0,
				Height:    0,
			},
		); err != nil {
			return nil, err
		}

		uploads = append(uploads, GalleryPhotoUploadTarget{
			PhotoID:   photoID,
			PutURL:    putURL.String(),
			ObjectKey: objectKey,
		})
	}

	return uploads, nil
}

func (s *Service) CompleteGalleryPhotoUpload(
	ctx context.Context,
	input CompleteGalleryPhotoInput,
) (galleriesrepo.GalleryPhoto, error) {
	if input.GalleryID == uuid.Nil {
		return galleriesrepo.GalleryPhoto{}, fmt.Errorf("gallery id is required")
	}

	if input.PhotoID == uuid.Nil {
		return galleriesrepo.GalleryPhoto{}, fmt.Errorf("photo id is required")
	}

	if input.PhotographerID == uuid.Nil {
		return galleriesrepo.GalleryPhoto{}, fmt.Errorf("photographer id is required")
	}

	photo, err := s.repo.GetPhotoForOwner(
		ctx,
		input.PhotoID,
		input.GalleryID,
		input.PhotographerID,
	)
	if err != nil {
		return galleriesrepo.GalleryPhoto{}, err
	}

	if _, err := s.storage.StatObject(ctx, photo.ImageKey); err != nil {
		return galleriesrepo.GalleryPhoto{}, err
	}

	reader, err := s.storage.GetObject(ctx, photo.ImageKey)
	if err != nil {
		return galleriesrepo.GalleryPhoto{}, err
	}
	defer reader.Close()

	config, _, err := image.DecodeConfig(reader)
	if err != nil {
		return galleriesrepo.GalleryPhoto{}, fmt.Errorf("decode gallery image config: %w", err)
	}

	completedPhoto, err := s.repo.MarkPhotoCompleted(
		ctx,
		input.PhotoID,
		input.GalleryID,
		input.PhotographerID,
		int32(config.Width),
		int32(config.Height),
	)
	if err != nil {
		return galleriesrepo.GalleryPhoto{}, err
	}

	return completedPhoto, nil
}

func buildGalleryPhotoObjectKey(
	galleryID uuid.UUID,
	photoID uuid.UUID,
	ext string,
) string {
	return fmt.Sprintf(
		"galleries/%s/%s%s",
		galleryID.String(),
		photoID.String(),
		ext,
	)
}
