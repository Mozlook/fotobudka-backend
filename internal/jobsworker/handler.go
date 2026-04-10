package jobsworker

import (
	"context"
	"errors"
	"fmt"

	"github.com/Mozlook/fotobudka-backend/internal/repository/jobs"
	"github.com/Mozlook/fotobudka-backend/internal/sessionphotos"
)

var ErrRetryableJob = errors.New("retryable job error")

func (w *Worker) handleJob(ctx context.Context, job jobs.Job) error {
	switch job.Type {
	case sessionphotos.JobTypeGenerateSessionPhotoVariants:
		if err := w.handleGenerateSessionPhotoVariants(ctx, job); err != nil {
			return fmt.Errorf("generate session photo variants: %w", err)
		}
		return nil

	default:
		return fmt.Errorf("unknown job type: %s", job.Type)
	}
}

func retryable(err error) error {
	if err == nil {
		return nil
	}
	return errors.Join(ErrRetryableJob, err)
}
