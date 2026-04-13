package middleware

import (
	"net/http"

	appauth "github.com/Mozlook/fotobudka-backend/internal/auth"
)

func RequireClientSessionAccess(manger appauth.ClientManager, next http.Handler) {
}
