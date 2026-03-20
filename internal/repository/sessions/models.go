package sessions

import "github.com/google/uuid"

type SessionOwner struct {
	ID             uuid.UUID
	PhotographerID uuid.UUID
}
