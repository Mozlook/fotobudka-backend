package finalphotos

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	dbgen "github.com/Mozlook/fotobudka-backend/internal/platform/db/sqlc"
	"github.com/Mozlook/fotobudka-backend/internal/platform/storage"
	"github.com/Mozlook/fotobudka-backend/internal/sessionphotos"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type FileInput struct {
	PhotoID   uuid.UUID
	Filename  string
	MimeType  string
	SizeBytes *int64
}

type FinalPutURL struct {
	FinalID   uuid.UUID
	PhotoID   uuid.UUID
	PutURL    string
	ObjectKey string
}

const presignedPutTTL = 30 * time.Minute

func (s *Service) PrepareFinalPhotoUploads(ctx context.Context, sessionID uuid.UUID, files []FileInput) ([]FinalPutURL, error) {
	if sessionID == uuid.Nil {
		return nil, ErrInvalidSessionID
	}
	if len(files) == 0 {
		return nil, ErrEmptyFiles
	}

	seen := make(map[uuid.UUID]struct{}, len(files))
	for _, file := range files {
		if file.PhotoID == uuid.Nil {
			return nil, ErrInvalidPhotoID
		}
		if _, exists := seen[file.PhotoID]; exists {
			return nil, fmt.Errorf("%w: %s", ErrDuplicatePhotoInBatch, file.PhotoID)
		}
		seen[file.PhotoID] = struct{}{}

		if strings.TrimSpace(file.Filename) == "" {
			return nil, ErrInvalidFilename
		}
		if strings.TrimSpace(file.MimeType) == "" {
			return nil, ErrInvalidMimeType
		}
		if file.SizeBytes == nil || *file.SizeBytes <= 0 {
			return nil, ErrInvalidFileSize
		}
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	qtx := dbgen.New(tx)

	sessionStatus, err := qtx.GetSessionStatusForUpdate(ctx, sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("get session status for update: %w", err)
	}

	if !AllowsFinalEditingOrDeliveryGeneration(sessionStatus.Status) {
		return nil, ErrFinalUploadLocked
	}

	finalURLList := make([]FinalPutURL, 0, len(files))

	for _, file := range files {
		selected, err := qtx.IsSelectedSessionPhoto(ctx, dbgen.IsSelectedSessionPhotoParams{
			SessionID: sessionID,
			PhotoID:   file.PhotoID,
		})
		if err != nil {
			return nil, fmt.Errorf("check selected photo %s: %w", file.PhotoID, err)
		}
		if !selected {
			return nil, fmt.Errorf("%w: %s", ErrPhotoNotSelected, file.PhotoID)
		}

		hasFinal, err := qtx.HasFinalPhotoForSessionPhoto(ctx, dbgen.HasFinalPhotoForSessionPhotoParams{
			SessionID: sessionID,
			PhotoID:   file.PhotoID,
		})
		if err != nil {
			return nil, fmt.Errorf("check existing final for photo %s: %w", file.PhotoID, err)
		}
		if hasFinal {
			return nil, fmt.Errorf("%w: %s", ErrFinalAlreadyExists, file.PhotoID)
		}

		ext, ok := sessionphotos.SourceExtFromMIME(file.MimeType)
		if !ok {
			return nil, ErrInvalidMimeType
		}

		finalID := uuid.New()
		objectKey := fmt.Sprintf("sessions/%s/final/%s%s", sessionID, finalID, ext)

		putURL, err := s.storage.PresignedPutObject(ctx, objectKey, presignedPutTTL)
		if err != nil {
			return nil, fmt.Errorf("presign put object for photo %s: %w", file.PhotoID, err)
		}
		if putURL == nil {
			return nil, fmt.Errorf("presigned put url is nil")
		}

		if err := qtx.InsertFinalPhoto(ctx, dbgen.InsertFinalPhotoParams{
			ID:             finalID,
			SessionID:      sessionID,
			PhotoID:        file.PhotoID,
			FinalKey:       objectKey,
			FinalSizeBytes: file.SizeBytes,
		}); err != nil {
			return nil, fmt.Errorf("insert final photo for photo %s: %w", file.PhotoID, err)
		}

		finalURLList = append(finalURLList, FinalPutURL{
			FinalID:   finalID,
			PhotoID:   file.PhotoID,
			PutURL:    putURL.String(),
			ObjectKey: objectKey,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return finalURLList, nil
}

func (s *Service) CompleteFinalPhotoUpload(ctx context.Context, sessionID, finalID uuid.UUID) error {
	if sessionID == uuid.Nil {
		return ErrInvalidSessionID
	}
	if finalID == uuid.Nil {
		return ErrInvalidFinalID
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	qtx := dbgen.New(tx)

	session, err := qtx.GetSessionStatusForUpdate(ctx, sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrSessionNotFound
		}
		return fmt.Errorf("get session status for update: %w", err)
	}

	if !AllowsFinalEditingOrDeliveryGeneration(session.Status) {
		return ErrFinalUploadLocked
	}

	finalPhoto, err := qtx.GetFinalPhotoByIDAndSessionID(ctx, dbgen.GetFinalPhotoByIDAndSessionIDParams{
		ID:        finalID,
		SessionID: sessionID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrFinalPhotoNotFound
		}
		return fmt.Errorf("get final photo by id and session id: %w", err)
	}

	if finalPhoto.FinalKey == "" {
		return fmt.Errorf("final key is empty")
	}

	object, err := s.storage.StatObject(ctx, finalPhoto.FinalKey)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotFound) {
			return ErrUploadedObjectNotFound
		}
		return fmt.Errorf("stat final object: %w", err)
	}

	size := object.Size
	rows, err := qtx.UpdateFinalPhotoSize(ctx, dbgen.UpdateFinalPhotoSizeParams{
		ID:             finalID,
		SessionID:      sessionID,
		FinalSizeBytes: &size,
	})
	if err != nil {
		return fmt.Errorf("update final photo size: %w", err)
	}
	if rows != 1 {
		return fmt.Errorf("update final photo size: unexpected affected rows: %d", rows)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
