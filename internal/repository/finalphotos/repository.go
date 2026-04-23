package finalphotos

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

func (r *Repository) ListFinalPhotosForDelivery(ctx context.Context, sessionID uuid.UUID) ([]FinalPhotoForDelivery, error) {
	rows, err := r.q.ListFinalPhotosForDelivery(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list final photos for delivery: %w", err)
	}

	finals := make([]FinalPhotoForDelivery, 0, len(rows))
	for _, row := range rows {
		finals = append(finals, FinalPhotoForDelivery{
			ID:               row.ID,
			PhotoID:          row.PhotoID,
			FinalKey:         row.FinalKey,
			OriginalFilename: row.OriginalFilename,
		})
	}

	return finals, nil
}
