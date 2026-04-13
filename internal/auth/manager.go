package appauth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Mozlook/fotobudka-backend/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

// ErrNoAuthToken is returned when the authentication cookie is missing
// or does not contain a token value.
var ErrNoAuthToken = errors.New("auth token cookie is missing")

// Manager signs and validates JWT tokens and manages authentication cookies.
type Manager struct {
	secret       []byte
	issuer       string
	audience     string
	ttl          time.Duration
	cookieName   string
	cookieDomain string
	cookieSecure bool
}

// NewManager creates a manager configured from application settings
func NewManager(cfg config.Config) *Manager {
	return &Manager{
		secret:       []byte(cfg.JWT.Secret),
		issuer:       cfg.JWT.Issuer,
		audience:     cfg.JWT.Audience,
		ttl:          time.Duration(cfg.JWT.TTLHours) * time.Hour,
		cookieName:   cfg.Cookie.Name,
		cookieDomain: cfg.Cookie.Domain,
		cookieSecure: cfg.Cookie.Secure,
	}
}

// ErrNoAuthToken is returned when the authentication cookie is missing
// or does not contain a token value.
func (m *Manager) IssueToken(userID string) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(m.ttl)

	claims := Claims{RegisteredClaims: jwt.RegisteredClaims{
		Subject:   userID,
		Issuer:    m.issuer,
		Audience:  jwt.ClaimStrings{m.audience},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// ParseAndValidate validates the provided token string and returns
// the authenticated user ID stored in the token subject claim.
func (m *Manager) ParseAndValidate(tokenString string) (string, error) {
	claims := &Claims{}

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
		return "", err
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	if claims.Subject == "" {
		return "", fmt.Errorf("token subject is empty")
	}
	return claims.Subject, nil
}

// SetAuthCookie writes the authentication cookie to the HTTP response.
func (m *Manager) SetAuthCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	cookie := &http.Cookie{
		Name:     m.cookieName,
		Value:    token,
		Path:     "/",
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

// ClearAuthCookie removes the authentication cookie from the client
func (m *Manager) ClearAuthCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     m.cookieName,
		Value:    "",
		Path:     "/",
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

// TokenFromRequest reads the authentication token from the request cookie
func (m *Manager) TokenFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie(m.cookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", ErrNoAuthToken
		}
		return "", err
	}

	if cookie.Value == "" {
		return "", ErrNoAuthToken
	}

	return cookie.Value, nil
}
