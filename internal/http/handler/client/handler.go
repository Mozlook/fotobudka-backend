package client

import (
	appauth "github.com/Mozlook/fotobudka-backend/internal/auth"
	"github.com/Mozlook/fotobudka-backend/internal/platform/redis"
	"github.com/Mozlook/fotobudka-backend/internal/sessionaccess"
	"github.com/Mozlook/fotobudka-backend/internal/sessionphotos"
)

// Handler serves /api/client endpoints.
type Handler struct {
	sessionPhotos      *sessionphotos.Service
	sessionAccess      *sessionaccess.Service
	redis              *redis.Client
	recaptchaSecretKey string
	clientManager      *appauth.ClientManager
}

// NewHandler creates a client handler
func NewHandler(
	sessionPhotos *sessionphotos.Service,
	sessionAccess *sessionaccess.Service,
	redis *redis.Client,
	recaptchaSecretKey string,
	clientManager *appauth.ClientManager,
) *Handler {
	return &Handler{
		sessionPhotos:      sessionPhotos,
		sessionAccess:      sessionAccess,
		redis:              redis,
		recaptchaSecretKey: recaptchaSecretKey,
		clientManager:      clientManager,
	}
}
