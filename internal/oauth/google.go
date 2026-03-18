package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Mozlook/fotobudka-backend/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleUserData contains the basic user profile returned by the Google
// OpenID Connect userinfo endpoint.
type GoogleUserData struct {
	// Sub is the stable Google account identifier.
	Sub string `json:"sub"`

	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// Provider wraps the OAuth 2.0 configuration used for Google authentication.
type Provider struct {
	config           oauth2.Config
	userInfoEndpoint string
}

// New creates a Google OAuth provider configured from application settings.
func New(cfg config.Config) *Provider {
	return &Provider{
		config: oauth2.Config{
			ClientID:     cfg.OAuth.GoogleClientID,
			ClientSecret: cfg.OAuth.GoogleClientSecret,
			RedirectURL:  cfg.OAuth.GoogleRedirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
		userInfoEndpoint: "https://openidconnect.googleapis.com/v1/userinfo",
	}
}

// LoginURL returns the Google authorization URL for the current login flow.
//
// The returned URL includes the provided state value and PKCE challenge.
func (p *Provider) LoginURL(state, verifier string) string {
	return p.config.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))
}

func (p *Provider) Exchange(ctx context.Context, code string, verifier string) (*oauth2.Token, error) {
	token, err := p.config.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return &oauth2.Token{}, err
	}
	return token, nil
}

func (p *Provider) FetchUserInfo(ctx context.Context, token *oauth2.Token) (GoogleUserData, error) {
	client := p.config.Client(ctx, token)

	res, err := client.Get(p.userInfoEndpoint)
	if err != nil {
		return GoogleUserData{}, fmt.Errorf("fetch google userinfo: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return GoogleUserData{}, fmt.Errorf(
			"google userinfo returned unexpected status: %d %s",
			res.StatusCode,
			http.StatusText(res.StatusCode),
		)
	}

	var userData GoogleUserData

	if err := json.NewDecoder(res.Body).Decode(&userData); err != nil {
		return GoogleUserData{}, fmt.Errorf("decode google userinfo response: %w", err)
	}

	return userData, nil
}
