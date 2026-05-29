package galleries

import (
	galleriesusecase "github.com/Mozlook/fotobudka-backend/internal/galleries"
	"github.com/Mozlook/fotobudka-backend/internal/platform/storage"
	galleriesrepo "github.com/Mozlook/fotobudka-backend/internal/repository/galleries"
)

type Handler struct {
	repo    *galleriesrepo.Repository
	service *galleriesusecase.Service
	storage *storage.Client
}

func New(
	repo *galleriesrepo.Repository,
	service *galleriesusecase.Service,
	storageClient *storage.Client,
) *Handler {
	return &Handler{
		repo:    repo,
		service: service,
		storage: storageClient,
	}
}
