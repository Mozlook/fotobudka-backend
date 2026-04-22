package deliveries

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	dbgen "github.com/Mozlook/fotobudka-backend/internal/platform/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const JobTypeGenerateDeliveryZIP = "generate_delivery_zip"

type GenerateDeliveryZIPPayload struct {
	SessionID  uuid.UUID `json:"session_id"`
	DeliveryID uuid.UUID `json:"delivery_id"`
}

type GenerateZIPResult struct {
	DeliveryID uuid.UUID `json:"delivery_id"`
	Version    int32     `json:"version"`
	Status     string    `json:"status"`
}

func (s *Service) GenerateZIP(ctx context.Context, sessionID uuid.UUID) (GenerateZIPResult, error) {
	if sessionID == uuid.Nil {
		return GenerateZIPResult{}, ErrInvalidSessionID
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return GenerateZIPResult{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	qtx := dbgen.New(tx)

	session, err := qtx.GetSessionStatusForUpdate(ctx, sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return GenerateZIPResult{}, ErrSessionNotFound
		}
		return GenerateZIPResult{}, fmt.Errorf("get session status for update: %w", err)
	}

	if session.Status != "editing" {
		return GenerateZIPResult{}, ErrGenerateZIPLocked
	}

	selectedCount, err := qtx.CountSelectionsBySessionID(ctx, sessionID)
	if err != nil {
		return GenerateZIPResult{}, fmt.Errorf("count selections by session id: %w", err)
	}
	if selectedCount == 0 {
		return GenerateZIPResult{}, ErrNoSelections
	}

	missingFinalsCount, err := qtx.CountSelectedPhotosWithoutFinal(ctx, sessionID)
	if err != nil {
		return GenerateZIPResult{}, fmt.Errorf("count selected photos without final: %w", err)
	}
	if missingFinalsCount > 0 {
		return GenerateZIPResult{}, ErrMissingFinalPhotos
	}

	nextVersion, err := qtx.GetNextDeliveryVersionForSession(ctx, sessionID)
	if err != nil {
		return GenerateZIPResult{}, fmt.Errorf("get next delivery version for session: %w", err)
	}

	deliveryID := uuid.New()

	if err := qtx.InsertDelivery(ctx, dbgen.InsertDeliveryParams{
		ID:        deliveryID,
		SessionID: sessionID,
		Version:   nextVersion,
		Status:    "generating",
	}); err != nil {
		return GenerateZIPResult{}, fmt.Errorf("insert delivery: %w", err)
	}

	payload, err := json.Marshal(GenerateDeliveryZIPPayload{
		SessionID:  sessionID,
		DeliveryID: deliveryID,
	})
	if err != nil {
		return GenerateZIPResult{}, fmt.Errorf("marshal generate delivery zip payload: %w", err)
	}

	if err := qtx.EnqueueJob(ctx, dbgen.EnqueueJobParams{
		ID:          uuid.New(),
		Type:        JobTypeGenerateDeliveryZIP,
		Payload:     payload,
		MaxAttempts: 3,
	}); err != nil {
		return GenerateZIPResult{}, fmt.Errorf("enqueue generate delivery zip job: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return GenerateZIPResult{}, fmt.Errorf("commit transaction: %w", err)
	}

	return GenerateZIPResult{
		DeliveryID: deliveryID,
		Version:    nextVersion,
		Status:     "generating",
	}, nil
}
