package sessions

import (
	sessionsrepo "github.com/Mozlook/fotobudka-backend/internal/repository/sessions"
	"github.com/Mozlook/fotobudka-backend/internal/sessionaccess"
)

// Handler serves authenticated /api/sessions endpoints.
type Handler struct {
	sessions       *sessionsrepo.Repository
	sessionAccess  *sessionaccess.Service
	frontendOrigin string
}

// NewHandler creates a sessions handler backed by the sessions repository.
func NewHandler(sessions *sessionsrepo.Repository, sessionAccess *sessionaccess.Service, frontendOrigin string) *Handler {
	return &Handler{sessions: sessions, sessionAccess: sessionAccess, frontendOrigin: frontendOrigin}
}
