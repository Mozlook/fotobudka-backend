package jobsworker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Mozlook/fotobudka-backend/internal/platform/storage"
	"github.com/Mozlook/fotobudka-backend/internal/repository/jobs"
	sessionphotosrepo "github.com/Mozlook/fotobudka-backend/internal/repository/sessionphotos"
	"github.com/google/uuid"
)

type Worker struct {
	workerID          string
	jobsRepo          *jobs.Repository
	sessionphotosRepo *sessionphotosrepo.Repository
	storage           *storage.Client
	limit             int32
}

func New(jobsRepo *jobs.Repository, sessionphotosRepo *sessionphotosrepo.Repository, storage *storage.Client, limit int32) *Worker {
	if limit <= 0 {
		limit = 10
	}
	return &Worker{
		workerID:          uuid.NewString(),
		jobsRepo:          jobsRepo,
		sessionphotosRepo: sessionphotosRepo,
		storage:           storage,
		limit:             limit,
	}
}

func (w *Worker) RunOnce(ctx context.Context) error {
	jobsToRun, err := w.jobsRepo.ClaimDueJobs(ctx, w.limit, w.workerID)
	if err != nil {
		return fmt.Errorf("claim due jobs: %w", err)
	}

	for _, job := range jobsToRun {
		jobErr := w.handleJob(ctx, job)
		if jobErr != nil {
			if errors.Is(jobErr, ErrRetryableJob) && job.Attempts < job.MaxAttempts {
				nextRunAt := w.nextRetryTime(job.Attempts)

				markErr := w.jobsRepo.MarkJobRetry(ctx, job.ID, jobErr.Error(), nextRunAt)
				if markErr != nil {
					return fmt.Errorf(
						"handle job %s failed: %v; additionally failed to mark retry: %w",
						job.ID,
						jobErr,
						markErr,
					)
				}

				continue
			}

			finalizeErr := w.handleFinalJobFailure(ctx, job, jobErr)

			markErr := w.jobsRepo.MarkJobFailed(ctx, job.ID, jobErr.Error())

			if finalizeErr != nil && markErr != nil {
				return fmt.Errorf(
					"handle job %s failed: %v; additionally failed to finalize job failure: %v; and failed to mark job failed: %w",
					job.ID,
					jobErr,
					finalizeErr,
					markErr,
				)
			}
			if finalizeErr != nil {
				return fmt.Errorf(
					"handle job %s failed: %v; additionally failed to finalize job failure: %w",
					job.ID,
					jobErr,
					finalizeErr,
				)
			}
			if markErr != nil {
				return fmt.Errorf(
					"handle job %s failed: %v; additionally failed to mark job failed: %w",
					job.ID,
					jobErr,
					markErr,
				)
			}

			continue
		}

		if err := w.jobsRepo.MarkJobSucceeded(ctx, job.ID); err != nil {
			return fmt.Errorf("mark job %s succeeded: %w", job.ID, err)
		}
	}

	return nil
}

func (w *Worker) nextRetryTime(attempts int32) time.Time {
	switch attempts {
	case 1:
		return time.Now().UTC().Add(30 * time.Second)
	case 2:
		return time.Now().UTC().Add(2 * time.Minute)
	default:
		return time.Now().UTC().Add(5 * time.Minute)
	}
}
