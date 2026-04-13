package middleware

import (
	"errors"
	"net/http"

	appauth "github.com/Mozlook/fotobudka-backend/internal/auth"
	sessionsrepo "github.com/Mozlook/fotobudka-backend/internal/repository/sessions"
	"github.com/google/uuid"
)

func RequireClientSessionAccess(manager *appauth.ClientManager, sessionAccessRepo *sessionsrepo.Repository, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := manager.TokenFromRequest(r)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		sessionAccessIDString, sessionIDString, err := manager.ParseAndValidateClient(token)
		if err != nil {
			manager.ClearClientCookie(w)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		sessionAccessID, err := uuid.Parse(sessionAccessIDString)
		if err != nil {
			manager.ClearClientCookie(w)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		sessionID, err := uuid.Parse(sessionIDString)
		if err != nil {
			manager.ClearClientCookie(w)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		sessionAccess, err := sessionAccessRepo.GetActiveClientSessionAccessByID(r.Context(), sessionAccessID)
		if err != nil {
			if errors.Is(err, sessionsrepo.ErrActiveClientSessionAccessNotFound) {

				manager.ClearClientCookie(w)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if sessionID != sessionAccess.SessionID {
			manager.ClearClientCookie(w)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		routeSessionIDString := r.PathValue("sessionId")
		if routeSessionIDString != "" {

			routeSessionID, err := uuid.Parse(routeSessionIDString)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			if routeSessionID != sessionID {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
		}

		ctx := clientContextWithSessionAccessID(r.Context(), sessionAccessID)
		ctx = clientContextWithSessionID(ctx, sessionID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
