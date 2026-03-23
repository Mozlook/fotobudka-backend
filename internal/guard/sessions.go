package guard

import (
	"context"
	"errors"

	"github.com/Mozlook/fotobudka-backend/internal/repository/sessions"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ErrSessionNotAccessible is returned whhen a session oes not exist
// or does not belong to the authenticated photographer.
var ErrSessionNotAccessible = errors.New("session not accessible")

// EnsureSessionOwner verifies that the given session exists
// and belongs to the provided authenticated user.
func EnsureSessionOwner(ctx context.Context, repo *sessions.Repository, sessionID, userID uuid.UUID) error {
	owner, err := repo.GetSessionOwnerByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrSessionNotAccessible
		}
		return err
	}

	if owner.PhotographerID != userID {
		return ErrSessionNotAccessible
	}
	return nil
}
