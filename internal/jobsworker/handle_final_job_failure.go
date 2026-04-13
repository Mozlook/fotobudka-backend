package jobsworker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Mozlook/fotobudka-backend/internal/repository/jobs"
	"github.com/Mozlook/fotobudka-backend/internal/sessionphotos"
	"github.com/google/uuid"
)

func (w *Worker) handleFinalJobFailure(ctx context.Context, job jobs.Job, cause error) error {
	_ = cause

	switch job.Type {
	case sessionphotos.JobTypeGenerateSessionPhotoVariants:
		var payload sessionphotos.GenerateSessionPhotoVariantsPayload
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return fmt.Errorf("unmarshal generate_session_photo_variants payload: %w", err)
		}

		sessionID := payload.SessionID
		photoID := payload.PhotoID

		if sessionID == uuid.Nil {
			return fmt.Errorf("sessionID cannot be empty")
		}
		if photoID == uuid.Nil {
			return fmt.Errorf("photoID cannot be empty")
		}

		if err := w.sessionPhotosRepo.MarkPhotoFailed(ctx, photoID, sessionID); err != nil {
			return fmt.Errorf("mark photo failed: %w", err)
		}

		err := w.reconcileSessionStatus(ctx, sessionID)
		if err != nil {
			// TODO: Add warning log
		}

		return nil

	default:
		return nil
	}
}
