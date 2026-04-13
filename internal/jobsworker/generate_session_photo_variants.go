package jobsworker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Mozlook/fotobudka-backend/internal/platform/storage"
	"github.com/Mozlook/fotobudka-backend/internal/repository/jobs"
	sessionphotosrepo "github.com/Mozlook/fotobudka-backend/internal/repository/sessionphotos"
	"github.com/Mozlook/fotobudka-backend/internal/sessionphotos"
	"github.com/google/uuid"
)

func (w *Worker) handleGenerateSessionPhotoVariants(ctx context.Context, job jobs.Job) error {
	var payload sessionphotos.GenerateSessionPhotoVariantsPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal generate_session_photo_variants payload: %w", err)
	}

	sessionID := payload.SessionID
	photoID := payload.PhotoID
	sourceKey := payload.SourceKey

	if sessionID == uuid.Nil {
		return fmt.Errorf("sessionID cannot be empty")
	}
	if photoID == uuid.Nil {
		return fmt.Errorf("photoID cannot be empty")
	}
	if sourceKey == "" {
		return fmt.Errorf("sourceKey cannot be empty")
	}

	if err := w.sessionPhotosRepo.MarkPhotoProcessing(ctx, photoID, sessionID); err != nil {
		if errors.Is(err, sessionphotosrepo.ErrSessionPhotoNotFound) {
			return fmt.Errorf("mark photo processing: %w", err)
		}
		return retryable(fmt.Errorf("mark photo processing: %w", err))
	}

	sourceImage, err := w.storage.GetObject(ctx, sourceKey)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotFound) {
			return fmt.Errorf("get object: %w", err)
		}
		return retryable(fmt.Errorf("get object: %w", err))
	}
	defer sourceImage.Close()

	srcImg, _, err := decodeImage(sourceImage)
	if err != nil {
		return fmt.Errorf("decode image: %w", err)
	}

	thumbImage := resizeToFit(srcImg, 400, 400)
	proofImage := resizeToFit(srcImg, 2000, 2000)

	proofWatermark, err := applyProofWatermark(proofImage, payload.WatermarkSeed)
	if err != nil {
		return fmt.Errorf("apply proof watermark: %w", err)
	}

	var thumbBuf bytes.Buffer
	if err := encodeJPEG(&thumbBuf, thumbImage, 82); err != nil {
		return fmt.Errorf("encode thumb jpeg: %w", err)
	}

	var proofBuf bytes.Buffer
	if err := encodeJPEG(&proofBuf, proofWatermark, 88); err != nil {
		return fmt.Errorf("encode proof jpeg: %w", err)
	}

	thumbKey := thumbObjectKey(sessionID, photoID)
	proofKey := proofObjectKey(sessionID, photoID)

	if err := w.storage.PutObject(
		ctx,
		thumbKey,
		"image/jpeg",
		bytes.NewReader(thumbBuf.Bytes()),
		int64(thumbBuf.Len()),
	); err != nil {
		return retryable(fmt.Errorf("put thumb: %w", err))
	}

	if err := w.storage.PutObject(
		ctx,
		proofKey,
		"image/jpeg",
		bytes.NewReader(proofBuf.Bytes()),
		int64(proofBuf.Len()),
	); err != nil {
		return retryable(fmt.Errorf("put proof: %w", err))
	}

	if err := w.sessionPhotosRepo.MarkPhotoReady(ctx, photoID, sessionID, thumbKey, proofKey); err != nil {
		if errors.Is(err, sessionphotosrepo.ErrSessionPhotoNotFound) {
			return fmt.Errorf("mark photo ready: %w", err)
		}
		return retryable(fmt.Errorf("mark photo ready: %w", err))
	}

	return nil
}
