package sessions

import (
	"context"
	"errors"
	"fmt"

	dbgen "github.com/Mozlook/fotobudka-backend/internal/platform/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrActiveClientSessionAccessNotFound = errors.New("active client session access not found")

type MiddlewareSessionAccess struct {
	SessionID       uuid.UUID
	SessionAccessID uuid.UUID
}

// InsertSessionAccess creates a new session access record for the given
// session and returns its identifier together with the creation timestamp.
func (r *Repository) InsertSessionAccess(ctx context.Context, in InsertSessionAccessInput) (SessionAccess, error) {
	row, err := r.q.InsertSessionAccess(ctx, dbgen.InsertSessionAccessParams{
		ID:        in.ID,
		SessionID: in.SessionID,
		CodeHmac:  in.CodeHmac,
		TokenHmac: in.TokenHmac,
	})
	if err != nil {
		return SessionAccess{}, fmt.Errorf("insert session access: %w", err)
	}
	return SessionAccess{ID: row.ID, CreatedAt: row.CreatedAt}, nil
}

// RevokeSessionAccess revokes all active access records for the given
// session and returns the revoked records.
//
// Only records that were active at the time of the update should be
// affected by the underlying query.
func (r *Repository) RevokeSessionAccess(ctx context.Context, sessionAccessID uuid.UUID) ([]RevokedSessionAccess, error) {
	rows, err := r.q.RevokeSessionAccess(ctx, sessionAccessID)
	if err != nil {
		return nil, fmt.Errorf("revoke session access: %w", err)
	}

	revokedSessionAccessList := make([]RevokedSessionAccess, 0, len(rows))

	for _, row := range rows {
		revokedSessionAccess := RevokedSessionAccess{
			ID:         row.ID,
			SessionID:  row.SessionID,
			CreatedAt:  row.CreatedAt,
			RevokedAt:  row.RevokedAt,
			LastUsedAt: row.LastUsedAt,
		}

		revokedSessionAccessList = append(revokedSessionAccessList, revokedSessionAccess)
	}
	return revokedSessionAccessList, nil
}

func (r *Repository) GetClientSessionByTokenHMAC(ctx context.Context, tokenHMAC string) (ClientSession, error) {
	row, err := r.q.GetClientSessionByTokenHMAC(ctx, tokenHMAC)
	if err != nil {
		return ClientSession{}, fmt.Errorf("get client session by tokenHMAC: %w", err)
	}

	return ClientSession{
		ID:              row.ID,
		Status:          row.Status,
		BasePriceCents:  row.BasePriceCents,
		IncludedCount:   row.IncludedCount,
		ExtraPriceCents: row.ExtraPriceCents,
		MinSelectCount:  row.MinSelectCount,
		Currency:        row.Currency,
		PaymentMode:     row.PaymentMode,
		Title:           row.Title,
	}, nil
}

func (r *Repository) GetClientSessionByCodeHMAC(ctx context.Context, codeHMAC string) (ClientSession, error) {
	row, err := r.q.GetClientSessionByCodeHMAC(ctx, codeHMAC)
	if err != nil {
		return ClientSession{}, fmt.Errorf("get client session by codeHMAC: %w", err)
	}

	return ClientSession{
		ID:              row.ID,
		Status:          row.Status,
		BasePriceCents:  row.BasePriceCents,
		IncludedCount:   row.IncludedCount,
		ExtraPriceCents: row.ExtraPriceCents,
		MinSelectCount:  row.MinSelectCount,
		Currency:        row.Currency,
		PaymentMode:     row.PaymentMode,
		Title:           row.Title,
	}, nil
}

func (r *Repository) GetActiveClientSessionAccessByID(ctx context.Context, sessionID uuid.UUID) (MiddlewareSessionAccess, error) {
	row, err := r.q.GetActiveClientSessionAccessByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return MiddlewareSessionAccess{}, ErrActiveClientSessionAccessNotFound
		}
		return MiddlewareSessionAccess{}, fmt.Errorf("get active client session access by id: %w", err)
	}
	return MiddlewareSessionAccess{SessionID: row.SessionID, SessionAccessID: row.ID}, nil
}
