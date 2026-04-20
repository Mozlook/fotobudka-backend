package finalphotos

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	dbgen "github.com/Mozlook/fotobudka-backend/internal/platform/db/sqlc"
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

	if sessionStatus.Status != "editing" {
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
