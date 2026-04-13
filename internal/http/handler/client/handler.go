package client

import (
	appauth "github.com/Mozlook/fotobudka-backend/internal/auth"
	"github.com/Mozlook/fotobudka-backend/internal/platform/redis"
	sessionphotosrepo "github.com/Mozlook/fotobudka-backend/internal/repository/sessionphotos"
	"github.com/Mozlook/fotobudka-backend/internal/sessionaccess"
)

// Handler serves /api/client endpoints.
type Handler struct {
	sessionPhotosRepo  *sessionphotosrepo.Repository
	sessionAccess      *sessionaccess.Service
	redis              *redis.Client
	recaptchaSecretKey string
	clientManager      *appauth.ClientManager
}

// NewHandler creates a client handler
func NewHandler(
	sessionPhotosRepo *sessionphotosrepo.Repository,
	sessionAccess *sessionaccess.Service,
	redis *redis.Client,
	recaptchaSecretKey string,
	clientManager *appauth.ClientManager,
) *Handler {
	return &Handler{
		sessionPhotosRepo:  sessionPhotosRepo,
		sessionAccess:      sessionAccess,
		redis:              redis,
		recaptchaSecretKey: recaptchaSecretKey,
		clientManager:      clientManager,
	}
}
