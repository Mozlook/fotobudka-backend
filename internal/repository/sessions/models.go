package sessions

import (
	"time"

	"github.com/google/uuid"
)

// SessionOwner represents the minimal ownership data needed to authorize
// access to a session resource.
type SessionOwner struct {
	ID             uuid.UUID
	PhotographerID uuid.UUID
}

// InsertSessionInput contains the data required to create a new session.
type InsertSessionInput struct {
	PhotographerID  uuid.UUID `json:"photographer_id"`
	Title           string    `json:"title"`
	ClientEmail     *string   `json:"client_email"`
	BasePriceCents  int32     `json:"base_price_cents"`
	IncludedCount   int32     `json:"included_count"`
	ExtraPriceCents int32     `json:"extra_price_cents"`
	MinSelectCount  int32     `json:"min_select_count"`
	Currency        string    `json:"currency"`
	PaymentMode     string    `json:"payment_mode"`
}

// GetSessionInput contains the filters used to list sessions for a photographer.
type GetSessionInput struct {
	PhotographerID uuid.UUID
	Offset         int32
}

// SessionStatus contains the identifier and current status of a session.
type SessionStatus struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

// Session represents a photographer session returned from the repository.
type Session struct {
	ID              uuid.UUID
	PhotographerID  uuid.UUID
	Title           string
	ClientEmail     *string
	Status          string
	BasePriceCents  int32
	IncludedCount   int32
	ExtraPriceCents int32
	MinSelectCount  int32
	Currency        string
	PaymentMode     string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ClosedAt        *time.Time
	DeleteAfter     *time.Time
}
