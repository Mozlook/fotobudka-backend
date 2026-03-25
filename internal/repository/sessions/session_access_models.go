package sessions

import (
	"time"

	"github.com/google/uuid"
)

// InsertSessionAccessInput contains the data required to create a new
// session access record.
type InsertSessionAccessInput struct {
	ID        uuid.UUID
	SessionID uuid.UUID
	CodeHmac  string
	TokenHmac string
}

// SessionAccess contains the minimal data returned after creating a
// session access record.
type SessionAccess struct {
	ID        uuid.UUID
	CreatedAt time.Time
}

// RevokedSessionAccess represents a session access record that has been
// revoked.
type RevokedSessionAccess struct {
	ID         uuid.UUID
	SessionID  uuid.UUID
	CreatedAt  time.Time
	RevokedAt  *time.Time
	LastUsedAt *time.Time
}
