package sessionphotos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	dbgen "github.com/Mozlook/fotobudka-backend/internal/platform/db/sqlc"
	"github.com/Mozlook/fotobudka-backend/internal/platform/storage"
	sessionphotosrepo "github.com/Mozlook/fotobudka-backend/internal/repository/sessionphotos"
	"github.com/google/uuid"
)

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
