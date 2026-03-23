package sessions

import sessionsrepo "github.com/Mozlook/fotobudka-backend/internal/repository/sessions"

// Handler serves authenticated /api/sessions endpoints.
type Handler struct {
	sessions *sessionsrepo.Repository
}

// NewHandler creates a sessions handler backed by the sessions repository.
func NewHandler(sessions *sessionsrepo.Repository) *Handler {
	return &Handler{sessions: sessions}
}
