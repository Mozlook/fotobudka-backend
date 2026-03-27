package client

import (
	"github.com/Mozlook/fotobudka-backend/internal/sessionaccess"
)

// Handler serves /api/client endpoints.
type Handler struct {
	sessionAccess *sessionaccess.Service
}

// NewHandler creates a client handler
func NewHandler(sessionAccess *sessionaccess.Service) *Handler {
	return &Handler{sessionAccess: sessionAccess}
}
