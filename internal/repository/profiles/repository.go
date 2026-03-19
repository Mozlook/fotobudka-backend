package profiles

import (
	"context"
	"encoding/json"
	"fmt"

	dbgen "github.com/Mozlook/fotobudka-backend/internal/platform/db/sqlc"
	"github.com/google/uuid"
)

// Repository provides persistence operations for photographer profiles.
type Repository struct {
	q *dbgen.Queries
}

// New creates a profiles repository backed by generated sqlc queries.
func New(q *dbgen.Queries) *Repository {
	return &Repository{q: q}
}

// GetPhotographerProfileByUserID returns the photographer profile assigned to the given user ID.
func (r *Repository) GetPhotographerProfileByUserID(ctx context.Context, userID uuid.UUID) (Profile, error) {
	row, err := r.q.GetPhotographerProfileByUserID(ctx, userID)
	if err != nil {
		return Profile{}, fmt.Errorf("get photographer profile by user id: %w", err)
	}

	var socialLinks SocialLinks
	if err := json.Unmarshal(row.SocialLinks, &socialLinks); err != nil {
		return Profile{}, fmt.Errorf("unmarshal social links: %w", err)
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

// UpsertPhotographerProfile creates or updates a photographer profile for the given user.
//
// When no profile exists for the provided user ID, a new record is inserted.
// When a profile already exists, the username, display name, bio, social links,
// and update timestamp are refreshed. The final persisted profile is returned.
func (r *Repository) UpsertPhotographerProfile(ctx context.Context, in UpsertInput) (Profile, error) {
	socialLinksJSON, err := json.Marshal(in.SocialLinks)
	if err != nil {
		return Profile{}, fmt.Errorf("marshal social links: %w", err)
	}

	row, err := r.q.UpsertPhotographerProfile(ctx, dbgen.UpsertPhotographerProfileParams{
		UserID:      in.UserID,
		Username:    in.Username,
		DisplayName: in.DisplayName,
		Bio:         in.Bio,
		SocialLinks: socialLinksJSON,
	})
	if err != nil {
		return Profile{}, fmt.Errorf("upsert photographer profile: %w", err)
	}

	var socialLinks SocialLinks
	if err := json.Unmarshal(row.SocialLinks, &socialLinks); err != nil {
		return Profile{}, fmt.Errorf("unmarshal social links: %w", err)
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
