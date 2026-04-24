package deliveries

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrDeliveryNotFound                = errors.New("delivery not found")
	ErrDeliveryNotFoundOrNotGenerating = errors.New("delivery not found or not generating")
	ErrLatestReadyDeliveryNotFound     = errors.New("latest ready delivery not found")
	ErrSessionDeliveryTransitionFailed = errors.New("session could not transition to delivered")
)

type Delivery struct {
	ID           uuid.UUID
	SessionID    uuid.UUID
	Version      int32
	Status       string
	ZipKey       *string
	ZipSizeBytes *int64
	CreatedAt    time.Time
	GeneratedAt  *time.Time
}

type LatestReadyDelivery struct {
	ID           uuid.UUID
	SessionID    uuid.UUID
	Version      int32
	ZipKey       string
	ZipSizeBytes *int64
	GeneratedAt  *time.Time
}
