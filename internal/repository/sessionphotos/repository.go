package sessionphotosrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	dbgen "github.com/Mozlook/fotobudka-backend/internal/platform/db/sqlc"
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
	sessionPhotosRepo *dbgen.Queries
	pool              *pgxpool.Pool
}

type SessionPhoto struct {
	ID              uuid.UUID
	SessionID       uuid.UUID
	SourceKey       string
	Status          string
	SourceSizeBytes int64
}

var ErrSessionPhotoNotFound = errors.New("session photo not found")

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

func (r *Repository) GetSessionPhotoByIDAndSessionID(ctx context.Context, photoID, sessionID uuid.UUID) (SessionPhoto, error) {
	photo, err := r.sessionPhotosRepo.GetSessionPhotoByIDAndSessionID(ctx, dbgen.GetSessionPhotoByIDAndSessionIDParams{
		ID:        photoID,
		SessionID: sessionID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SessionPhoto{}, ErrSessionPhotoNotFound
		}
		return SessionPhoto{}, fmt.Errorf("get session photo by id and session_id: %w", err)
	}

	return SessionPhoto{
		ID:              photo.ID,
		SessionID:       photo.SessionID,
		SourceKey:       photo.SourceKey,
		Status:          photo.Status,
		SourceSizeBytes: photo.SourceSizeBytes,
	}, nil
}

func (r *Repository) MarkSessionPhotoUploaded(ctx context.Context, photoID, sessionID uuid.UUID, sourceSizeBytes int64) error {
	count, err := r.sessionPhotosRepo.MarkSessionPhotoUploaded(ctx, dbgen.MarkSessionPhotoUploadedParams{
		ID:              photoID,
		SessionID:       sessionID,
		SourceSizeBytes: sourceSizeBytes,
	})
	if err != nil {
		return fmt.Errorf("mark session photo uploaded: %w", err)
	}

	if count == 0 {
		return ErrSessionPhotoNotFound
	}
	if count != 1 {
		return fmt.Errorf("mark session photo uploaded: unexpected affected rows: %d", count)
	}

	return nil
}
