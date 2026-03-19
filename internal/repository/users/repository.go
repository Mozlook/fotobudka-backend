package users

import (
	"context"
	"fmt"

	dbgen "github.com/Mozlook/fotobudka-backend/internal/platform/db/sqlc"
)

// Repository provides persistence operations for users.
type Repository struct {
	q *dbgen.Queries
}

// New creates a users repository backed by generated sqlc queries.
func New(q *dbgen.Queries) *Repository {
	return &Repository{q: q}
}

// UpsertFromGoogle creates or updates a user based on the Google account identifier.
//
// When no user exists for the provided GoogleSub, a new record is inserted.
// When a matching user already exists, the email, name, and avatar URL are updated.
// The final persisted user record is returned.
func (r *Repository) UpsertFromGoogle(ctx context.Context, in UpsertFromGoogleInput) (User, error) {
	row, err := r.q.UpsertUserFromGoogle(ctx, dbgen.UpsertUserFromGoogleParams{
		ID:        in.ID,
		GoogleSub: in.GoogleSub,
		Email:     in.Email,
		Name:      in.Name,
		AvatarUrl: in.AvatarURL,
	})
	if err != nil {
		return User{}, fmt.Errorf("upsert user from google:%w", err)
	}

	return User{
		ID:        row.ID,
		GoogleSub: row.GoogleSub,
		Email:     row.Email,
		Name:      row.Name,
		AvatarURL: row.AvatarUrl,
		CreatedAt: row.CreatedAt,
	}, nil
}
