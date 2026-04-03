package sessionphotosrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InsertPhotoRow struct {
	ID               uuid.UUID
	SessionID        uuid.UUID
	OriginalFilename string
	MimeType         string
	SourceKey        string
	SourceSizeBytes  int64
	Status           string
	WatermarkSeed    int32
	CreatedAt        time.Time
}

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) InsertBatch(ctx context.Context, rows []InsertPhotoRow) (int64, error) {
	if len(rows) == 0 {
		return 0, nil
	}

	count, err := r.pool.CopyFrom(
		ctx,
		pgx.Identifier{"session_photos"},
		[]string{
			"id",
			"session_id",
			"original_filename",
			"mime_type",
			"source_key",
			"source_size_bytes",
			"status",
			"watermark_seed",
			"created_at",
		},
		pgx.CopyFromSlice(len(rows), func(i int) ([]any, error) {
			row := rows[i]

			return []any{
				row.ID,
				row.SessionID,
				row.OriginalFilename,
				row.MimeType,
				row.SourceKey,
				row.SourceSizeBytes,
				row.Status,
				row.WatermarkSeed,
				row.CreatedAt,
			}, nil
		}),
	)
	if err != nil {
		return 0, fmt.Errorf("copy session_photos: %w", err)
	}

	return count, nil
}
