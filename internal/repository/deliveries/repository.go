package deliveries

import (
	"context"
	"errors"
	"fmt"

	dbgen "github.com/Mozlook/fotobudka-backend/internal/platform/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	q    *dbgen.Queries
	pool *pgxpool.Pool
}

func New(q *dbgen.Queries, pool *pgxpool.Pool) *Repository {
	return &Repository{
		q:    q,
		pool: pool,
	}
}

func (r *Repository) GetDeliveryByID(ctx context.Context, deliveryID uuid.UUID) (Delivery, error) {
	row, err := r.q.GetDeliveryByID(ctx, deliveryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Delivery{}, ErrDeliveryNotFound
		}
		return Delivery{}, fmt.Errorf("get delivery by id: %w", err)
	}

	return Delivery{
		ID:           row.ID,
		SessionID:    row.SessionID,
		Version:      row.Version,
		Status:       row.Status,
		ZipKey:       row.ZipKey,
		ZipSizeBytes: row.ZipSizeBytes,
		CreatedAt:    row.CreatedAt,
		GeneratedAt:  row.GeneratedAt,
	}, nil
}

func (r *Repository) MarkDeliveryFailed(ctx context.Context, deliveryID uuid.UUID) error {
	rows, err := r.q.MarkDeliveryFailed(ctx, deliveryID)
	if err != nil {
		return fmt.Errorf("mark delivery failed: %w", err)
	}

	if rows == 0 {
		return ErrDeliveryNotFoundOrNotGenerating
	}
	if rows != 1 {
		return fmt.Errorf("mark delivery failed: unexpected affected rows: %d", rows)
	}

	return nil
}

func (r *Repository) MarkDeliveryReadyAndSessionDelivered(
	ctx context.Context,
	deliveryID, sessionID uuid.UUID,
	zipKey string,
	zipSizeBytes int64,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	qtx := dbgen.New(tx)

	rows, err := qtx.MarkDeliveryReady(ctx, dbgen.MarkDeliveryReadyParams{
		ID:           deliveryID,
		ZipKey:       &zipKey,
		ZipSizeBytes: &zipSizeBytes,
	})
	if err != nil {
		return fmt.Errorf("mark delivery ready: %w", err)
	}
	if rows == 0 {
		return ErrDeliveryNotFoundOrNotGenerating
	}
	if rows != 1 {
		return fmt.Errorf("mark delivery ready: unexpected affected rows: %d", rows)
	}

	sessionRows, err := qtx.MarkSessionDelivered(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("mark session delivered: %w", err)
	}
	if sessionRows == 0 {
		return ErrSessionDeliveryTransitionFailed
	}
	if sessionRows != 1 {
		return fmt.Errorf("mark session delivered: unexpected affected rows: %d", sessionRows)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
