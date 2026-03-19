package users

import (
	"time"

	"github.com/google/uuid"
)

type UpsertFromGoogleInput struct {
	ID        uuid.UUID
	GoogleSub string
	Email     string
	Name      string
	AvatarURL string
}

type User struct {
	ID        uuid.UUID
	GoogleSub string
	Email     string
	Name      string
	AvatarURL *string
	CreatedAt time.Time
}
