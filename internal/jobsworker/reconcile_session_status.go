package jobsworker

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (w *Worker) reconcileSessionStatus(ctx context.Context, sessionID uuid.UUID) error {
	stats, err := w.sessionPhotosRepo.GetSessionPhotoStats(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("get session by id: %w", err)
	}
	unfinished := stats.pending_upload + stats.UploadedCount + stats.ProcessingCount

	if unfinished > 0 {
		return nil
	}

	if stats.ReadyCount > 0 {
		w.sessionsRepo.TryMarkSessionSelecting(ctx, sessionID)
		return nil
	}

	if stats.FailedCount > 0 {
		w.sessionsRepo.TryMarkSessionFailed(ctx, sessionID)
		return nil
	}

	return nil
}
