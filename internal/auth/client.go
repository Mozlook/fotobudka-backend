package appauth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Mozlook/fotobudka-backend/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrNoClientToken = errors.New("auth client cookie is missing")

type ClientManager struct {
	secret       []byte
	issuer       string
	audience     string
	ttl          time.Duration
	cookieName   string
	cookieDomain string
	cookieSecure bool
}

func NewClientManager(cfg config.Config) *ClientManager {
	return &ClientManager{
		secret:       []byte(cfg.JWT.Secret),
		issuer:       cfg.JWT.Issuer,
		audience:     cfg.JWT.ClientAudience,
		ttl:          24 * time.Hour,
		cookieName:   cfg.Cookie.ClientName,
		cookieDomain: cfg.Cookie.Domain,
		cookieSecure: cfg.Cookie.Secure,
	}
}

func (m *ClientManager) IssueClientToken(sessionAccessID, sessionID uuid.UUID) (string, time.Time, error) {
	if sessionAccessID == uuid.Nil {
		return "", time.Time{}, fmt.Errorf("sessionAccessID cannot be empty")
	}
	if sessionID == uuid.Nil {
		return "", time.Time{}, fmt.Errorf("sessionID cannot be empty")
	}

	now := time.Now().UTC()
	expiresAt := now.Add(m.ttl)

	claims := ClientClaims{
		SessionID: sessionID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sessionAccessID.String(),
			Issuer:    m.issuer,
			Audience:  jwt.ClaimStrings{m.audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

func (m *ClientManager) ParseAndValidateClient(tokenString string) (string, string, error) {
	claims := &ClientClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return m.secret, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(m.issuer),
		jwt.WithAudience(m.audience),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)
	if err != nil {
		return "", "", fmt.Errorf("parse client token: %w", err)
	}

	if !token.Valid {
		return "", "", fmt.Errorf("invalid token")
	}

	if claims.Subject == "" {
		return "", "", fmt.Errorf("token subject is empty")
	}
	if claims.SessionID == "" {
		return "", "", fmt.Errorf("token sessionID is empty")
	}
	return claims.Subject, claims.SessionID, nil
}

func (m *ClientManager) SetClientCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	cookie := &http.Cookie{
		Name:     m.cookieName,
		Value:    token,
		Path:     "/api/client",
		Expires:  expiresAt.UTC(),
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		Secure:   m.cookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	if m.cookieDomain != "" {
		cookie.Domain = m.cookieDomain
	}

	http.SetCookie(w, cookie)
}

func (m *ClientManager) ClearClientCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     m.cookieName,
		Value:    "",
		Path:     "/api/client",
		Expires:  time.Unix(0, 0).UTC(),
		MaxAge:   -1,
		Secure:   m.cookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	if m.cookieDomain != "" {
		cookie.Domain = m.cookieDomain
	}

	http.SetCookie(w, cookie)
}

func (m *ClientManager) TokenFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie(m.cookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", ErrNoAuthToken
		}
		return "", err
	}

	if cookie.Value == "" {
		return "", ErrNoClientToken
	}

	return cookie.Value, nil
}
