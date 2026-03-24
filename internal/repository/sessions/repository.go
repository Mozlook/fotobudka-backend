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

// GetSessions returns a paginated list of sessions owned by the given
// photographer.
func (r *Repository) GetSessions(ctx context.Context, in GetSessionsInput) ([]Session, error) {
	queryResponse, err := r.q.GetSessions(ctx, dbgen.GetSessionsParams{
		PhotographerID: in.PhotographerID,
		Offset:         in.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("get sessions: %w", err)
	}

	sessionsList := make([]Session, 0, len(queryResponse))

	for _, row := range queryResponse {
		session := Session{
			ID:              row.ID,
			PhotographerID:  row.PhotographerID,
			Title:           row.Title,
			ClientEmail:     row.ClientEmail,
			Status:          row.Status,
			BasePriceCents:  row.BasePriceCents,
			IncludedCount:   row.IncludedCount,
			ExtraPriceCents: row.ExtraPriceCents,
			MinSelectCount:  row.MinSelectCount,
			Currency:        row.Currency,
			PaymentMode:     row.PaymentMode,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
			ClosedAt:        row.ClosedAt,
			DeleteAfter:     row.DeleteAfter,
		}

		sessionsList = append(sessionsList, session)
	}

	return sessionsList, nil
}

// GetSessionByID returns the full session details for the given session ID.
func (r *Repository) GetSessionByID(ctx context.Context, id uuid.UUID) (Session, error) {
	row, err := r.q.GetSessionByID(ctx, id)
	if err != nil {
		return Session{}, fmt.Errorf("get sessions: %w", err)
	}

	session := Session{
		ID:              row.ID,
		PhotographerID:  row.PhotographerID,
		Title:           row.Title,
		ClientEmail:     row.ClientEmail,
		Status:          row.Status,
		BasePriceCents:  row.BasePriceCents,
		IncludedCount:   row.IncludedCount,
		ExtraPriceCents: row.ExtraPriceCents,
		MinSelectCount:  row.MinSelectCount,
		Currency:        row.Currency,
		PaymentMode:     row.PaymentMode,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		ClosedAt:        row.ClosedAt,
		DeleteAfter:     row.DeleteAfter,
	}
	return session, nil
}

func (r *Repository) CloseSession(ctx context.Context, sessionID uuid.UUID) (ClosedSession, error) {
	row, err := r.q.CloseSession(ctx, sessionID)
	if err != nil {
		return ClosedSession{}, fmt.Errorf("close session: %w", err)
	}

	return ClosedSession{
		ID:          row.ID,
		Title:       row.Title,
		Status:      string,
		ClosedAt:    row.ClosedAt,
		DeleteAfter: row.DeleteAfter,
	}, nil
}
