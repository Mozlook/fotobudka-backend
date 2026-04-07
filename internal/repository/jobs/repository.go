package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	dbgen "github.com/Mozlook/fotobudka-backend/internal/platform/db/sqlc"
	"github.com/google/uuid"
)

type Job struct {
	ID          uuid.UUID
	Type        string
	Status      string
	Payload     []byte
	Attempts    int32
	MaxAttempts int32
	NextRunAt   time.Time
	LockedAt    *time.Time
	LockedBy    *string
	LastError   *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Repository provides access to jobs table jobs table jobs table jobs table jobs table jobs table jobs table jobs table operations.
type Repository struct {
	q *dbgen.Queries
}

// New creates a new Repository backed by sqlc queries.
func New(q *dbgen.Queries) *Repository {
	return &Repository{q: q}
}

type EnqueueJobInput struct {
	ID          uuid.UUID
	Type        string
	Payload     []byte
	MaxAttempts int32
}

var ErrJobNotFoundOrNotRunning = errors.New("job not found or not running")

func (r *Repository) EnqueueJob(ctx context.Context, in EnqueueJobInput) error {
	err := r.q.EnqueueJob(ctx, dbgen.EnqueueJobParams{
		ID:          in.ID,
		Type:        in.Type,
		Payload:     in.Payload,
		MaxAttempts: in.MaxAttempts,
	})
	if err != nil {
		return fmt.Errorf("job enqueue: %w", err)
	}

	return nil
}

func (r *Repository) ClaimDueJobs(ctx context.Context, limit int32, lockedBy string) ([]Job, error) {
	rows, err := r.q.ClaimDueJobs(ctx, dbgen.ClaimDueJobsParams{
		LockedBy:   &lockedBy,
		LimitCount: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("claim due jobs: %w", err)
	}

	jobs := make([]Job, 0, len(rows))

	for _, row := range rows {
		job := Job{
			ID:          row.ID,
			Type:        row.Type,
			Status:      row.Status,
			Payload:     row.Payload,
			Attempts:    row.Attempts,
			MaxAttempts: row.MaxAttempts,
			NextRunAt:   row.NextRunAt,
			LockedAt:    row.LockedAt,
			LockedBy:    row.LockedBy,
			LastError:   row.LastError,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		}
		jobs = append(jobs, job)

	}
	return jobs, nil
}

func (r *Repository) MarkJobSucceeded(ctx context.Context, jobID uuid.UUID) error {
	rows, err := r.q.MarkJobSucceeded(ctx, jobID)
	if err != nil {
		return fmt.Errorf("mark job succeeded: %w", err)
	}

	if rows == 0 {
		return ErrJobNotFoundOrNotRunning
	}
	if rows != 1 {
		return fmt.Errorf("mark job succeeded: unexpected affected rows: %d", rows)
	}

	return nil
}

func (r *Repository) MarkJobRetry(ctx context.Context, jobID uuid.UUID, lastError string, nextRunAt time.Time) error {
	rows, err := r.q.MarkJobRetry(ctx, dbgen.MarkJobRetryParams{
		ID:        jobID,
		LastError: &lastError,
		NextRunAt: nextRunAt,
	})
	if err != nil {
		return fmt.Errorf("mark job retry: %w", err)
	}

	if rows == 0 {
		return ErrJobNotFoundOrNotRunning
	}
	if rows != 1 {
		return fmt.Errorf("mark job retry: unexpected affected rows: %d", rows)
	}

	return nil
}

func (r *Repository) MarkJobFailed(ctx context.Context, jobID uuid.UUID, lastError string) error {
	rows, err := r.q.MarkJobFailed(ctx, dbgen.MarkJobFailedParams{
		ID:        jobID,
		LastError: &lastError,
	})
	if err != nil {
		return fmt.Errorf("mark job failed: %w", err)
	}

	if rows == 0 {
		return ErrJobNotFoundOrNotRunning
	}
	if rows != 1 {
		return fmt.Errorf("mark job failed: unexpected affected rows: %d", rows)
	}

	return nil
}
