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
	WatermarkSeed   int32
}

type PhotoStats struct {
	PendingUploadCount int64
	TotalCount         int64
	UploadedCount      int64
	ProcessingCount    int64
	ReadyCount         int64
	FailedCount        int64
}

type ClientSessionPhoto struct {
	PhotoID  uuid.UUID
	ThumbKey *string
	Selected bool
	Note     *string
}

var ErrSessionPhotoNotFound = errors.New("session photo not found")

func New(sessionPhotosRepo *dbgen.Queries, pool *pgxpool.Pool) *Repository {
	return &Repository{
		sessionPhotosRepo: sessionPhotosRepo,
		pool:              pool,
	}
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
		WatermarkSeed:   photo.WatermarkSeed,
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

func (r *Repository) MarkPhotoProcessing(ctx context.Context, photoID, sessionID uuid.UUID) error {
	count, err := r.sessionPhotosRepo.MarkPhotoProcessing(ctx, dbgen.MarkPhotoProcessingParams{
		ID:        photoID,
		SessionID: sessionID,
	})
	if err != nil {
		return fmt.Errorf("mark session photo processing: %w", err)
	}

	if count == 0 {
		return ErrSessionPhotoNotFound
	}
	if count != 1 {
		return fmt.Errorf("mark session photo processing: unexpected affected rows: %d", count)
	}

	return nil
}

func (r *Repository) MarkPhotoReady(ctx context.Context, photoID, sessionID uuid.UUID, thumbKey, proofKey string) error {
	count, err := r.sessionPhotosRepo.MarkPhotoReady(ctx, dbgen.MarkPhotoReadyParams{
		ID:        photoID,
		SessionID: sessionID,
		ThumbKey:  &thumbKey,
		ProofKey:  &proofKey,
	})
	if err != nil {
		return fmt.Errorf("mark session photo ready: %w", err)
	}

	if count == 0 {
		return ErrSessionPhotoNotFound
	}
	if count != 1 {
		return fmt.Errorf("mark session photo ready: unexpected affected rows: %d", count)
	}

	return nil
}

func (r *Repository) MarkPhotoFailed(ctx context.Context, photoID, sessionID uuid.UUID) error {
	count, err := r.sessionPhotosRepo.MarkPhotoFailed(ctx, dbgen.MarkPhotoFailedParams{
		ID:        photoID,
		SessionID: sessionID,
	})
	if err != nil {
		return fmt.Errorf("mark session photo failed: %w", err)
	}

	if count == 0 {
		return ErrSessionPhotoNotFound
	}
	if count != 1 {
		return fmt.Errorf("mark session photo failed: unexpected affected rows: %d", count)
	}

	return nil
}

func (r *Repository) GetSessionPhotoStats(ctx context.Context, sessionID uuid.UUID) (PhotoStats, error) {
	stats, err := r.sessionPhotosRepo.GetSessionPhotoStats(ctx, sessionID)
	if err != nil {
		return PhotoStats{}, fmt.Errorf("get session photo stats: %w", err)
	}

	return PhotoStats{
		PendingUploadCount: stats.PendingUploadCount,
		TotalCount:         stats.TotalCount,
		UploadedCount:      stats.UploadedCount,
		ProcessingCount:    stats.ProcessingCount,
		ReadyCount:         stats.ReadyCount,
		FailedCount:        stats.FailedCount,
	}, nil
}

func (r *Repository) ListReadyClientSessionPhotos(ctx context.Context, sessionID uuid.UUID, offsetCount int32) ([]ClientSessionPhoto, error) {
	rows, err := r.sessionPhotosRepo.ListReadyClientSessionPhotos(ctx, dbgen.ListReadyClientSessionPhotosParams{
		SessionID:   sessionID,
		OffsetCount: offsetCount,
		LimitCount:  200,
	})
	if err != nil {
		return nil, fmt.Errorf("list ready client session photos: %w", err)
	}

	photos := make([]ClientSessionPhoto, 0, len(rows))
	for _, row := range rows {
		photos = append(photos, ClientSessionPhoto{
			PhotoID:  row.PhotoID,
			ThumbKey: row.ThumbKey,
			Selected: row.Selected,
			Note:     row.Note,
		})
	}

	return photos, nil
}

func (r *Repository) GetReadyClientPhotoProofKey(ctx context.Context, sessionID, photoID uuid.UUID) (string, error) {
	proofKey, err := r.sessionPhotosRepo.GetReadyClientPhotoProofKey(ctx, dbgen.GetReadyClientPhotoProofKeyParams{SessionID: sessionID, PhotoID: photoID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrSessionPhotoNotFound
		}
		return "", fmt.Errorf("get ready client photo proof key: %w", err)
	}

	return *proofKey, nil
}
