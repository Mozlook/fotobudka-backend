package deliveries

import (
	"github.com/Mozlook/fotobudka-backend/internal/platform/storage"
	deliveriesrepo "github.com/Mozlook/fotobudka-backend/internal/repository/deliveries"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool    *pgxpool.Pool
	repo    *deliveriesrepo.Repository
	storage *storage.Client
}

func New(pool *pgxpool.Pool, repo *deliveriesrepo.Repository, storageClient *storage.Client) *Service {
	return &Service{
		pool:    pool,
		repo:    repo,
		storage: storageClient,
	}
}
