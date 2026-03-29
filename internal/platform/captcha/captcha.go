package captcha

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type VerifyResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
}

func Verify(ctx context.Context, secret, captchaToken, remoteIP string) (bool, error) {
	if strings.TrimSpace(captchaToken) == "" {
		return false, nil
	}

	form := url.Values{}
	form.Set("secret", secret)
	form.Set("response", captchaToken)
	if strings.TrimSpace(remoteIP) != "" {
		form.Set("remoteip", remoteIP)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://www.google.com/recaptcha/api/siteverify",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return false, fmt.Errorf("build captcha verify request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("send captcha verify request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, fmt.Errorf("captcha verify returned status %d", resp.StatusCode)
	}

	var verifyResp VerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&verifyResp); err != nil {
		return false, fmt.Errorf("decode captcha verify response: %w", err)
	}

	return verifyResp.Success, nil
}
