package jobs

import (
	"context"
	"fmt"

	dbgen "github.com/Mozlook/fotobudka-backend/internal/platform/db/sqlc"
	"github.com/google/uuid"
)

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
