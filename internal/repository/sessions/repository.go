package sessions

import (
	"context"
	"fmt"

	dbgen "github.com/Mozlook/fotobudka-backend/internal/platform/db/sqlc"
	"github.com/google/uuid"
)

type Repository struct {
	q *dbgen.Queries
}

func New(q *dbgen.Queries) *Repository {
	return &Repository{q: q}
}

func (r *Repository) GetSessionOwnerByID(ctx context.Context, sessionID uuid.UUID) (SessionOwner, error) {
	sessionOwner, err := r.q.GetSessionOwnerByID(ctx, sessionID)
	if err != nil {
		return SessionOwner{}, fmt.Errorf("get session owner by id: %w", err)
	}
	return SessionOwner{
		ID:             sessionOwner.ID,
		PhotographerID: sessionOwner.PhotographerID,
	}, nil
}
