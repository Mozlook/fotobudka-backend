package sessionaccess

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"fmt"

	"github.com/Mozlook/fotobudka-backend/internal/repository/sessions"
	"github.com/google/uuid"
)

type PlainTextResponse struct {
	PlainTextToken string
	PlainTextCode  string
}

func GenerateAccess(sessionID uuid.UUID, secret []byte) (sessions.InsertSessionAccessInput, PlainTextResponse, error) {
	if len(secret) == 0 {
		return sessions.InsertSessionAccessInput{}, PlainTextResponse{}, fmt.Errorf("can't generate access with empty secret")
	}
	id := uuid.New()

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return sessions.InsertSessionAccessInput{}, PlainTextResponse{}, fmt.Errorf("generate token bytes: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	codeBytes := make([]byte, 5)
	if _, err := rand.Read(codeBytes); err != nil {
		return sessions.InsertSessionAccessInput{}, PlainTextResponse{}, fmt.Errorf("generate code bytes: %w", err)
	}
	code := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(codeBytes)

	tokenHmac := hmacHex(secret, token)
	codeHmac := hmacHex(secret, code)

	insertInput := sessions.InsertSessionAccessInput{
		ID:        id,
		SessionID: sessionID,
		TokenHmac: tokenHmac,
		CodeHmac:  codeHmac,
	}

	plainTextResponse := PlainTextResponse{
		PlainTextToken: token,
		PlainTextCode:  code,
	}

	return insertInput, plainTextResponse, nil
}

func hmacHex(secret []byte, value string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(value))
	sum := mac.Sum(nil)

	return hex.EncodeToString(sum)
}
