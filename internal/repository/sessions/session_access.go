package sessions

import (
	"context"
	"fmt"

	dbgen "github.com/Mozlook/fotobudka-backend/internal/platform/db/sqlc"
	"github.com/google/uuid"
)

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

func (r *Repository) RevokeSessionAccess(ctx context.Context, sessionID uuid.UUID) ([]RevokedSessionAccess, error) {
	rows, err := r.q.RevokeSessionAccess(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("revoke session access: %w", err)
	}

	revokedSessionAccessList := make([]RevokedSessionAccess, len(rows))

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
