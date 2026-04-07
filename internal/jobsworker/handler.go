package jobsworker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Mozlook/fotobudka-backend/internal/repository/jobs"
	"github.com/Mozlook/fotobudka-backend/internal/sessionphotos"
)

var ErrRetryableJob = errors.New("retryable job error")

func (w *Worker) handleJob(ctx context.Context, job jobs.Job) error {
	switch job.Type {
	case sessionphotos.JobTypeGenerateSessionPhotoVariants:
		var payload sessionphotos.GenerateSessionPhotoVariantsPayload
		err := json.Unmarshal(job.Payload, &payload)
		if err != nil {
			return fmt.Errorf("unmarshal generate_session_photo_variants payload: %w", err)
		}

		return fmt.Errorf("generate_session_photo_variants not implemented yet")

	default:
		return fmt.Errorf("unknown job type: %s", job.Type)
	}
}

func retryable(err error) error {
	return fmt.Errorf("%w: %s", ErrRetryableJob, err.Error())
}
