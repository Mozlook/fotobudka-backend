package selections

import (
	"github.com/Mozlook/fotobudka-backend/internal/repository/sessions"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool         *pgxpool.Pool
	sessionsRepo *sessions.Repository
}

func New(
	pool *pgxpool.Pool,
	sessionsRepo *sessions.Repository,
) *Service {
	return &Service{
		pool:         pool,
		sessionsRepo: sessionsRepo,
	}
}
