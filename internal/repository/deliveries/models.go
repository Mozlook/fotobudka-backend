package deliveries

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrDeliveryNotFound                = errors.New("delivery not found")
	ErrDeliveryNotFoundOrNotGenerating = errors.New("delivery not found or not generating")
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
