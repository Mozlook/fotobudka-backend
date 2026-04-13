package jobsworker

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (w *Worker) reconcileSessionStatus(ctx context.Context, sessionID uuid.UUID) error {
	stats, err := w.sessionPhotosRepo.GetSessionPhotoStats(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("get session photo stats: %w", err)
	}

	unfinished := stats.PendingUploadCount + stats.UploadedCount + stats.ProcessingCount

	if unfinished > 0 {
		return nil
	}

	if stats.ReadyCount > 0 {
		_, err := w.sessionsRepo.TryMarkSessionSelecting(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("mark session selecting: %w", err)
		}
		return nil
	}

	if stats.FailedCount > 0 {
		_, err := w.sessionsRepo.TryMarkSessionFailed(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("mark session failed: %w", err)
		}
		return nil
	}

	return nil
}
