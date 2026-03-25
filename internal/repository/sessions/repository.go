package sessions

import (
	dbgen "github.com/Mozlook/fotobudka-backend/internal/platform/db/sqlc"
)

// Repository provides access to session persistence operations.
type Repository struct {
	q *dbgen.Queries
}

// New creates a new Repository backed by sqlc queries.
func New(q *dbgen.Queries) *Repository {
	return &Repository{q: q}
}
