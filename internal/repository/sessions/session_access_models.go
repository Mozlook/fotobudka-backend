package sessions

import (
	"time"

	"github.com/google/uuid"
)

type InsertSessionAccessInput struct {
	ID        uuid.UUID
	SessionID uuid.UUID
	CodeHmac  string
	TokenHmac string
}

type SessionAccess struct {
	ID        uuid.UUID
	CreatedAt time.Time
}

type RevokedSessionAccess struct {
	ID         uuid.UUID
	SessionID  uuid.UUID
	CreatedAt  time.Time
	RevokedAt  *time.Time
	LastUsedAt *time.Time
}
