package sessions

import (
	"context"
	"fmt"

	dbgen "github.com/Mozlook/fotobudka-backend/internal/platform/db/sqlc"
	"github.com/google/uuid"
)

// Repository provides access to session persistence operations.
type Repository struct {
	q *dbgen.Queries
}

// New creates a new Repository backed by sqlc queries.
func New(q *dbgen.Queries) *Repository {
	return &Repository{q: q}
}

// GetSessionOwnerByID returns the session identifier and owner identifier
// for the given session.
//
// It is used by ownership guards to verify whether the authenticated
// photographer can access the requested session.
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

// InsertSession creates a new session for the given photographer and returns
// the created session identifier together with its initial status.
func (r *Repository) InsertSession(ctx context.Context, in InsertSessionInput) (SessionStatus, error) {
	session, err := r.q.InsertSession(ctx, dbgen.InsertSessionParams{
		PhotographerID:  in.PhotographerID,
		Title:           in.Title,
		ClientEmail:     in.ClientEmail,
		BasePriceCents:  in.BasePriceCents,
		IncludedCount:   in.IncludedCount,
		ExtraPriceCents: in.ExtraPriceCents,
		MinSelectCount:  in.MinSelectCount,
		Currency:        in.Currency,
		PaymentMode:     in.PaymentMode,
	})
	if err != nil {
		return SessionStatus{}, fmt.Errorf("insert session: %w", err)
	}
	return SessionStatus{
		ID:     session.ID,
		Status: session.Status,
	}, nil
}
