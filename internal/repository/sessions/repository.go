package sessions

import (
	dbgen "github.com/Mozlook/fotobudka-backend/internal/platform/db/sqlc"
	"github.com/jackc/pgx/v5"
)

// Repository provides access to session persistence operations.
type Repository struct {
	q *dbgen.Queries
}

// New creates a new Repository backed by sqlc queries.
func New(q *dbgen.Queries) *Repository {
	return &Repository{q: q}
}

// WithTx returns a repository bound to the provided transaction.
func (r *Repository) WithTx(tx pgx.Tx) *Repository {
	return &Repository{
		q: r.q.WithTx(tx),
	}
}
