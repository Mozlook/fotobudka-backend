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
	ID              uuid.UUID
	PhotographerID  uuid.UUID
	Title           string
	ClientEmail     *string
	BasePriceCents  int32
	IncludedCount   int32
	ExtraPriceCents int32
	MinSelectCount  int32
	Currency        string
	PaymentMode     string
}

// GetSessionsInput contains the filters used to list sessions for a photographer.
type GetSessionsInput struct {
	PhotographerID uuid.UUID
	Offset         int32
}

// SessionStatus contains the identifier and current status of a session.
type SessionStatus struct {
	ID     uuid.UUID
	Status string
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

// ClosedSession contains the data returned after a session has been
// marked as closed.
type ClosedSession struct {
	ID          uuid.UUID
	Title       string
	Status      string
	ClosedAt    *time.Time
	DeleteAfter *time.Time
}
