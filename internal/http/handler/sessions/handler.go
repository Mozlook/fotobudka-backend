package sessions

import (
	sessionsphotorepo "github.com/Mozlook/fotobudka-backend/internal/repository/sessionphotos"
	sessionsrepo "github.com/Mozlook/fotobudka-backend/internal/repository/sessions"
	"github.com/Mozlook/fotobudka-backend/internal/sessionaccess"
	"github.com/Mozlook/fotobudka-backend/internal/sessionphotos"
)

// Handler serves authenticated /api/sessions endpoints.
type Handler struct {
	sessionsRepo      *sessionsrepo.Repository
	sessionAccess     *sessionaccess.Service
	sessionPhotos     *sessionphotos.Service
	sessionsPhotoRepo *sessionsphotorepo.Repository
	frontendOrigin    string
}

// NewHandler creates a sessions handler backed by the sessions repository.
func NewHandler(
	sessions *sessionsrepo.Repository,
	sessionAccess *sessionaccess.Service,
	sessionPhotos *sessionphotos.Service,
	sessionsPhotoRepo *sessionsphotorepo.Repository,
	frontendOrigin string,
) *Handler {
	return &Handler{
		sessionsRepo:      sessions,
		sessionAccess:     sessionAccess,
		sessionPhotos:     sessionPhotos,
		sessionsPhotoRepo: sessionsPhotoRepo,
		frontendOrigin:    frontendOrigin,
	}
}
