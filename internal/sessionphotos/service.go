package sessionphotos

import (
	"context"
	"errors"
	"fmt"
	"hash/crc32"
	"net/url"
	"strings"
	"time"

	"github.com/Mozlook/fotobudka-backend/internal/platform/storage"
	sessionphotosrepo "github.com/Mozlook/fotobudka-backend/internal/repository/sessionphotos"
	"github.com/google/uuid"
)

const presignedPutTTL = 30 * time.Minute

type PhotoPutURL struct {
	PhotoID   uuid.UUID
	PutURL    *url.URL
	ObjectKey string
	Error     bool
}

type FileInput struct {
	Filename  string `json:"filename"`
	MimeType  string `json:"mime_type"`
	SizeBytes int64  `json:"size_bytes"`
}

type Service struct {
	storage *storage.Client
	repo    *sessionphotosrepo.Repository
}

var (
	ErrInvalidPhotoStatus     = errors.New("invalid photo status")
	ErrSessionPhotoNotFound   = errors.New("session photo not found")
	ErrUploadedObjectNotFound = errors.New("photo object not found")
)

func New(storageClient *storage.Client, repo *sessionphotosrepo.Repository) *Service {
	return &Service{
		storage: storageClient,
		repo:    repo,
	}
}

func (s *Service) presignedUploadURLs(ctx context.Context, sessionID string, files []FileInput) ([]PhotoPutURL, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, fmt.Errorf("session_id cannot be empty")
	}

	output := make([]PhotoPutURL, len(files))

	for i, file := range files {
		if strings.TrimSpace(file.Filename) == "" || file.SizeBytes <= 0 {
			output[i] = PhotoPutURL{Error: true}
			continue
		}

		ext, ok := sourceExtFromMIME(file.MimeType)
		if !ok {
			output[i] = PhotoPutURL{Error: true}
			continue
		}

		photoID := uuid.New()
		objectKey := fmt.Sprintf("sessions/%s/source/%v%s", sessionID, photoID, ext)

		putURL, err := s.storage.PresignedPutObject(ctx, objectKey, presignedPutTTL)
		if err != nil {
			return nil, fmt.Errorf("presign put object for %q: %w", file.Filename, err)
		}

		output[i] = PhotoPutURL{
			PhotoID:   photoID,
			PutURL:    putURL,
			ObjectKey: objectKey,
			Error:     false,
		}
	}

	return output, nil
}

func (s *Service) PrepareSessionPhotoUploads(ctx context.Context, sessionID uuid.UUID, files []FileInput) ([]PhotoPutURL, error) {
	if sessionID == uuid.Nil {
		return nil, fmt.Errorf("session_id cannot be nil")
	}
	urls, err := s.presignedUploadURLs(ctx, sessionID.String(), files)
	if err != nil {
		return nil, fmt.Errorf("prepare presigned upload urls: %w", err)
	}

	insertRows := make([]sessionphotosrepo.InsertPhotoRow, 0, len(files))

	now := time.Now().UTC()
	for i, upload := range urls {
		if upload.Error != false {
			continue
		}

		row := sessionphotosrepo.InsertPhotoRow{
			ID:               upload.PhotoID,
			SessionID:        sessionID,
			OriginalFilename: files[i].Filename,
			MimeType:         files[i].MimeType,
			SourceKey:        upload.ObjectKey,
			SourceSizeBytes:  files[i].SizeBytes,
			Status:           "pending_upload",
			WatermarkSeed:    watermarkSeedFromPhotoID(upload.PhotoID),
			CreatedAt:        now,
		}
		insertRows = append(insertRows, row)
	}

	_, err = s.repo.InsertBatch(ctx, insertRows)
	if err != nil {
		return nil, fmt.Errorf("insert session_photos batch: %w", err)
	}
	return urls, nil
}

func (s *Service) CompleteUpload(ctx context.Context, sessionID, photoID uuid.UUID) error {
	photo, err := s.repo.GetSessionPhotoByIDAndSessionID(ctx, photoID, sessionID)
	if err != nil {
		if errors.Is(err, sessionphotosrepo.ErrSessionPhotoNotFound) {
			return ErrSessionPhotoNotFound
		}
		return fmt.Errorf("get session photo: %w", err)
	}

	if photo.Status == "uploaded" {
		return nil
	}

	if photo.Status != "pending_upload" {
		return ErrInvalidPhotoStatus
	}

	object, err := s.storage.StatObject(ctx, photo.SourceKey)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotFound) {
			return ErrUploadedObjectNotFound
		}
		return fmt.Errorf("stat uploaded object: %w", err)
	}

	if err := s.repo.MarkSessionPhotoUploaded(ctx, photoID, sessionID, object.Size); err != nil {
		if errors.Is(err, sessionphotosrepo.ErrSessionPhotoNotFound) {
			return ErrSessionPhotoNotFound
		}
		return fmt.Errorf("mark session photo uploaded: %w", err)
	}

	return nil
}

func sourceExtFromMIME(mimeType string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "image/jpeg", "image/jpg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/webp":
		return ".webp", true
	default:
		return "", false
	}
}

func watermarkSeedFromPhotoID(id uuid.UUID) int32 {
	sum := crc32.ChecksumIEEE(id[:])
	return int32(sum & 0x7fffffff)
}
