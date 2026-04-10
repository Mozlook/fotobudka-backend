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
		err := w.handleGenerateSessionPhotoVariants(ctx, job)
		if err != nil {
			return fmt.Errorf("generate session photo variant: %w", err)
		}

	default:
		return fmt.Errorf("unknown job type: %s", job.Type)
	}
}

func retryable(err error) error {
	return fmt.Errorf("%w: %s", ErrRetryableJob, err.Error())
}
