package appauth

import (
	"fmt"
	"time"

	"github.com/Mozlook/fotobudka-backend/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

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
	if sessionAccessID.String() == "" {
		return "", time.Time{}, fmt.Errorf("sessionAccessID cannot be empty")
	}
	if sessionID.String() == "" {
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
