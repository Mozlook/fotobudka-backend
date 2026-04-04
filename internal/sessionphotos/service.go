package sessionphotos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"net/url"
	"strings"
	"time"

	dbgen "github.com/Mozlook/fotobudka-backend/internal/platform/db/sqlc"
	"github.com/Mozlook/fotobudka-backend/internal/platform/storage"
	"github.com/Mozlook/fotobudka-backend/internal/repository/jobs"
	sessionphotosrepo "github.com/Mozlook/fotobudka-backend/internal/repository/sessionphotos"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
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

const JobTypeGenerateSessionPhotoVariants = "generate_session_photo_variants"

type GenerateSessionPhotoVariantsPayload struct {
	SessionID     uuid.UUID `json:"session_id"`
	PhotoID       uuid.UUID `json:"photo_id"`
	SourceKey     string    `json:"source_key"`
	WatermarkSeed int32     `json:"watermark_seed"`
}

type Service struct {
	storage    *storage.Client
	photosRepo *sessionphotosrepo.Repository
	jobsRepo   *jobs.Repository
	pool       *pgxpool.Pool
}

var (
	ErrInvalidPhotoStatus     = errors.New("invalid photo status")
	ErrSessionPhotoNotFound   = errors.New("session photo not found")
	ErrUploadedObjectNotFound = errors.New("photo object not found")
)

func New(storageClient *storage.Client, photosRepo *sessionphotosrepo.Repository, jobsRepo *jobs.Repository, pool *pgxpool.Pool) *Service {
	return &Service{
		storage:    storageClient,
		photosRepo: photosRepo,
		jobsRepo:   jobsRepo,
		pool:       pool,
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

	_, err = s.photosRepo.InsertBatch(ctx, insertRows)
	if err != nil {
		return nil, fmt.Errorf("insert session_photos batch: %w", err)
	}
	return urls, nil
}

func (s *Service) CompleteUpload(ctx context.Context, sessionID, photoID uuid.UUID) error {
	photo, err := s.photosRepo.GetSessionPhotoByIDAndSessionID(ctx, photoID, sessionID)
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

	payload := GenerateSessionPhotoVariantsPayload{
		SessionID:     sessionID,
		PhotoID:       photoID,
		SourceKey:     photo.SourceKey,
		WatermarkSeed: photo.WatermarkSeed,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal generate_session_photo_variants payload: %w", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	qtx := dbgen.New(tx)

	rows, err := qtx.MarkSessionPhotoUploaded(ctx, dbgen.MarkSessionPhotoUploadedParams{
		ID:              photoID,
		SessionID:       sessionID,
		SourceSizeBytes: object.Size,
	})
	if err != nil {
		return fmt.Errorf("mark session photo uploaded: %w", err)
	}
	if rows != 1 {
		return fmt.Errorf("mark session photo uploaded: unexpected affected rows: %d", rows)
	}

	err = qtx.EnqueueJob(ctx, dbgen.EnqueueJobParams{
		ID:          uuid.New(),
		Type:        JobTypeGenerateSessionPhotoVariants,
		Payload:     payloadJSON,
		MaxAttempts: 3,
	})
	if err != nil {
		return fmt.Errorf("enqueue generate_session_photo_variants job: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
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
