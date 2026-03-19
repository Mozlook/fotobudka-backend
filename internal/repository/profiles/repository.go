package profiles

import (
	"context"
	"encoding/json"
	"fmt"

	dbgen "github.com/Mozlook/fotobudka-backend/internal/platform/db/sqlc"
	"github.com/google/uuid"
)

// Repository provides persistence operations for users.
type Repository struct {
	q *dbgen.Queries
}

// New creates a profiles repository backed by generated sqlc queries.
func New(q *dbgen.Queries) *Repository {
	return &Repository{q: q}
}

func (r *Repository) GetPhotographerProfileByUserID(ctx context.Context, userID uuid.UUID) (Profile, error) {
	row, err := r.q.GetPhotographerProfileByUserID(ctx, userID)
	if err != nil {
		return Profile{}, fmt.Errorf("get photographer profile by id: %w", err)
	}

	var socialLinks SocialLinks
	err = json.Unmarshal(row.SocialLinks, &socialLinks)
	if err != nil {
		return Profile{}, fmt.Errorf("socialLinks parsing error: %w", err)
	}

	return Profile{
		UserID:      row.UserID,
		Username:    row.Username,
		DisplayName: row.DisplayName,
		Bio:         row.Bio,
		SocialLinks: socialLinks,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}
