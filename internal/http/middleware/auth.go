package middleware

import (
	"context"
	"net/http"

	appauth "github.com/Mozlook/fotobudka-backend/internal/auth"
	"github.com/google/uuid"
)

const userIDKey contextKey = "user_id"

// RequireAuth ensures that the request contains a valid application auth token.
//
// It reads the auth cookie from the request, validates the JWT,
// parses the authenticated user ID, stores it in the request context,
// and forwards the request to the next handler.
func RequireAuth(manager *appauth.Manager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := manager.TokenFromRequest(r)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		userID, err := manager.ParseAndValidate(token)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		parsedUserID, err := uuid.Parse(userID)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, parsedUserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
