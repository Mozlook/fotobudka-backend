package users

import (
	"context"

	dbgen "github.com/Mozlook/fotobudka-backend/internal/platform/db/sqlc"
)

type Repository struct {
	q *dbgen.Queries
}

func New(q *dbgen.Queries) *Repository {
	return &Repository{q: q}
}

func (r *Repository) UpsertFromGoogle(ctx context.Context, in UpsertFromGoogleInput) (User, error) {
	row, err := r.q.UpsertUserFromGoogle(ctx, dbgen.UpsertUserFromGoogleParams{
		ID:        in.ID,
		GoogleSub: in.GoogleSub,
		Email:     in.Email,
		Name:      in.Name,
		AvatarUrl: in.AvatarURL,
	})
	if err != nil {
		return User{}, err
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
