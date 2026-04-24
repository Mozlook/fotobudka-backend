package deliveries

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type LatestDeliveryDownloadResult struct {
	DeliveryID   uuid.UUID  `json:"delivery_id"`
	Version      int32      `json:"version"`
	DownloadURL  string     `json:"download_url"`
	ZipSizeBytes *int64     `json:"zip_size_bytes,omitempty"`
	GeneratedAt  *time.Time `json:"generated_at,omitempty"`
}

const presignedDownloadTTL = 30 * time.Minute

func (s *Service) GetLatestDeliveryDownloadURL(ctx context.Context, sessionID uuid.UUID) (LatestDeliveryDownloadResult, error) {
	if sessionID == uuid.Nil {
		return LatestDeliveryDownloadResult{}, ErrInvalidSessionID
	}

	delivery, err := s.repo.GetLatestReadyDeliveryBySessionID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, ErrLatestReadyDeliveryNotFound) {
			return LatestDeliveryDownloadResult{}, ErrLatestReadyDeliveryNotFound
		}
		return LatestDeliveryDownloadResult{}, fmt.Errorf("get latest ready delivery by session id: %w", err)
	}

	downloadURL, err := s.storage.PresignedGetObject(ctx, delivery.ZipKey, presignedDownloadTTL)
	if err != nil {
		return LatestDeliveryDownloadResult{}, fmt.Errorf("presign delivery zip object: %w", err)
	}

	return LatestDeliveryDownloadResult{
		DeliveryID:   delivery.ID,
		Version:      delivery.Version,
		DownloadURL:  downloadURL,
		ZipSizeBytes: delivery.ZipSizeBytes,
		GeneratedAt:  delivery.GeneratedAt,
	}, nil
}
