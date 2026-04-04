package sessionphotos

import (
	"context"
	"fmt"
	"strings"
	"time"

	sessionphotosrepo "github.com/Mozlook/fotobudka-backend/internal/repository/sessionphotos"
	"github.com/google/uuid"
)

const presignedPutTTL = 30 * time.Minute

func (s *Service) presignUploadURLs(ctx context.Context, sessionID string, files []FileInput) ([]PhotoPutURL, error) {
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
		objectKey := fmt.Sprintf("sessions/%s/source/%s%s", sessionID, photoID, ext)

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

	urls, err := s.presignUploadURLs(ctx, sessionID.String(), files)
	if err != nil {
		return nil, fmt.Errorf("prepare presigned upload urls: %w", err)
	}

	insertRows := make([]sessionphotosrepo.InsertPhotoRow, 0, len(files))
	now := time.Now().UTC()

	for i, upload := range urls {
		if upload.Error {
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
