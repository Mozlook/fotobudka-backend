package sessionphotos

import (
	"github.com/Mozlook/fotobudka-backend/internal/platform/storage"
	sessionphotosrepo "github.com/Mozlook/fotobudka-backend/internal/repository/sessionphotos"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	storage    *storage.Client
	photosRepo *sessionphotosrepo.Repository
	pool       *pgxpool.Pool
}

func New(
	storageClient *storage.Client,
	photosRepo *sessionphotosrepo.Repository,
	pool *pgxpool.Pool,
) *Service {
	return &Service{
		storage:    storageClient,
		photosRepo: photosRepo,
		pool:       pool,
	}
}
