package finalphotos

import (
	"github.com/Mozlook/fotobudka-backend/internal/platform/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	storage *storage.Client
	pool    *pgxpool.Pool
}

func New(
	storageClient *storage.Client,
	pool *pgxpool.Pool,
) *Service {
	return &Service{
		storage: storageClient,
		pool:    pool,
	}
}
