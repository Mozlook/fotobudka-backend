package config

import (
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"strings"
)

// Validate checks whether the configuration is complete and internally consistent.
func (c Config) Validate() error {
	var errs []string

	if isBlank(c.App.Name) {
		errs = append(errs, "APP_NAME is required")
	}

	switch c.App.Env {
	case "dev", "test", "prod":
	default:
		errs = append(errs, "APP_ENV must be one of: dev, test, prod")
	}

	if isBlank(c.HTTP.APIAddr) {
		errs = append(errs, "API_ADDR is required")
	} else if err := validateListenAddr(c.HTTP.APIAddr); err != nil {
		errs = append(errs, fmt.Sprintf("API_ADDR %s", err.Error()))
	}

	if isBlank(c.HTTP.BaseURL) {
		errs = append(errs, "BASE_URL is required")
	} else if err := validateURL(c.HTTP.BaseURL, "http", "https"); err != nil {
		errs = append(errs, fmt.Sprintf("BASE_URL %s", err.Error()))
	}

	if isBlank(c.HTTP.FrontendOrigin) {
		errs = append(errs, "FRONTEND_ORIGIN is required")
	} else if err := validateOrigin(c.HTTP.FrontendOrigin); err != nil {
		errs = append(errs, fmt.Sprintf("FRONTEND_ORIGIN %s", err.Error()))
	}

	if isBlank(c.DB.URL) {
		errs = append(errs, "DB_URL is required")
	} else if err := validateURL(c.DB.URL, "postgres", "postgresql"); err != nil {
		errs = append(errs, fmt.Sprintf("DB_URL %s", err.Error()))
	}

	if isBlank(c.S3.Endpoint) {
		errs = append(errs, "S3_ENDPOINT is required")
	} else if err := validateS3Endpoint(c.S3.Endpoint); err != nil {
		errs = append(errs, fmt.Sprintf("S3_ENDPOINT %s", err.Error()))
	}

	if isBlank(c.S3.Bucket) {
		errs = append(errs, "S3_BUCKET is required")
	}

	if isBlank(c.S3.AccessKeyID) {
		errs = append(errs, "S3_ACCESS_KEY_ID is required")
	}

	if isBlank(c.S3.SecretAccessKey) {
		errs = append(errs, "S3_SECRET_ACCESS_KEY is required")
	}

	if isBlank(c.S3.Region) {
		errs = append(errs, "S3_REGION is required")
	}

	if isBlank(c.SIEM.LogDir) {
		errs = append(errs, "SIEM_LOG_DIR is required")
	}

	if hasAny(c.SMTP.Host, c.SMTP.User, c.SMTP.Password, c.SMTP.From) || c.SMTP.Port != 0 {
		if isBlank(c.SMTP.Host) {
			errs = append(errs, "SMTP_HOST is required when SMTP is configured")
		}

		if c.SMTP.Port <= 0 || c.SMTP.Port > 65535 {
			errs = append(errs, "SMTP_PORT must be between 1 and 65535 when SMTP is configured")
		}

		if !isBlank(c.SMTP.From) {
			if _, err := mail.ParseAddress(c.SMTP.From); err != nil {
				errs = append(errs, "SMTP_FROM must be a valid email address")
			}
		}

		if hasAny(c.SMTP.User, c.SMTP.Password) && !hasAll(c.SMTP.User, c.SMTP.Password) {
			errs = append(errs, "SMTP_USER and SMTP_PASSWORD must be set together")
		}
	}

	if hasAny(c.OAuth.GoogleClientID, c.OAuth.GoogleClientSecret) &&
		!hasAll(c.OAuth.GoogleClientID, c.OAuth.GoogleClientSecret) {
		errs = append(errs, "GOOGLE_OAUTH_CLIENT_ID and GOOGLE_OAUTH_CLIENT_SECRET must be set together")
	}

	if isBlank(c.OAuth.GoogleClientID) {
		errs = append(errs, "GOOGLE_OAUTH_CLIENT_ID is required")
	}

	if isBlank(c.OAuth.GoogleClientSecret) {
		errs = append(errs, "GOOGLE_OAUTH_CLIENT_SECRET is required")
	}

	if isBlank(c.OAuth.GoogleRedirectURL) {
		errs = append(errs, "GOOGLE_OAUTH_REDIRECT_URL is required")
	} else if err := validateURL(c.OAuth.GoogleRedirectURL, "http", "https"); err != nil {
		errs = append(errs, fmt.Sprintf("GOOGLE_OAUTH_REDIRECT_URL %s", err.Error()))
	}

	if hasAny(c.Captcha.RecaptchaSiteKey, c.Captcha.RecaptchaSecretKey) &&
		!hasAll(c.Captcha.RecaptchaSiteKey, c.Captcha.RecaptchaSecretKey) {
		errs = append(errs, "RECAPTCHA_SITE_KEY and RECAPTCHA_SECRET_KEY must be set together")
	}

	if c.Captcha.FailedCodeAttemptsTTL <= 0 {
		errs = append(errs, "CODE_LOGIN_ATTEMPTS_TTL must be greater than 0")
	}

	if c.Captcha.CodeCaptchaThreshold <= 0 {
		errs = append(errs, "CODE_LOGIN_CAPTCHA_THRESHOLD must be greater than 0")
	}

	if !isBlank(c.Redis.URL) {
		if err := validateURL(c.Redis.URL, "redis", "rediss"); err != nil {
			errs = append(errs, fmt.Sprintf("REDIS_URL %s", err.Error()))
		}
	}

	if (c.Captcha.FailedCodeAttemptsTTL > 0 || c.Captcha.CodeCaptchaThreshold > 0) && isBlank(c.Redis.URL) {
		errs = append(errs, "REDIS_URL is required when code login attempts protection is enabled")
	}

	if isBlank(c.JWT.Secret) {
		errs = append(errs, "JWT_SECRET is required")
	} else if len(c.JWT.Secret) < 32 {
		errs = append(errs, "JWT_SECRET must be at least 32 characters long")
	}

	if isBlank(c.JWT.Issuer) {
		errs = append(errs, "JWT_ISSUER is required")
	}

	if isBlank(c.JWT.Audience) {
		errs = append(errs, "JWT_AUDIENCE is required")
	}

	if c.JWT.TTLHours <= 0 {
		errs = append(errs, "JWT_TTL_HOURS must be greater than 0")
	}

	if isBlank(c.Cookie.Name) {
		errs = append(errs, "COOKIE_NAME is required")
	} else if err := validateCookieName(c.Cookie.Name); err != nil {
		errs = append(errs, fmt.Sprintf("COOKIE_NAME %s", err.Error()))
	}

	if !isBlank(c.Cookie.Domain) {
		if err := validateCookieDomain(c.Cookie.Domain); err != nil {
			errs = append(errs, fmt.Sprintf("COOKIE_DOMAIN %s", err.Error()))
		}
	}

	if strings.HasPrefix(c.Cookie.Name, "__Host-") {
		if !c.Cookie.Secure {
			errs = append(errs, "COOKIE_SECURE must be true when COOKIE_NAME starts with __Host-")
		}
		if !isBlank(c.Cookie.Domain) {
			errs = append(errs, "COOKIE_DOMAIN must be empty when COOKIE_NAME starts with __Host-")
		}
	}

	if c.App.Env == "prod" && !c.Cookie.Secure {
		errs = append(errs, "COOKIE_SECURE must be true in prod")
	}

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed: %s", strings.Join(errs, ", "))
	}

	return nil
}

// validateListenAddr checks whether addr is a valid listen address in host:port or :port form.
func validateListenAddr(addr string) error {
	_, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("must be in host:port or :port format")
	}

	return nil
}

// validateURL checks whether raw is a valid URL and whether it uses one of the allowed schemes.
func validateURL(raw string, allowedSchemes ...string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("is invalid: %w", err)
	}

	if parsed.Scheme == "" {
		return fmt.Errorf("must include a scheme")
	}

	if parsed.Host == "" {
		return fmt.Errorf("must include a host")
	}

	if len(allowedSchemes) > 0 {
		validScheme := false

		for _, scheme := range allowedSchemes {
			if strings.EqualFold(parsed.Scheme, scheme) {
				validScheme = true
				break
			}
		}

		if !validScheme {
			return fmt.Errorf("must use one of the schemes: %s", strings.Join(allowedSchemes, ", "))
		}
	}

	return nil
}

func validateOrigin(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("is invalid: %w", err)
	}

	if parsed.Scheme == "" {
		return fmt.Errorf("must include a scheme")
	}

	if parsed.Host == "" {
		return fmt.Errorf("must include a host")
	}

	if !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") {
		return fmt.Errorf("must use one of the schemes: http, https")
	}

	if parsed.Path != "" && parsed.Path != "/" {
		return fmt.Errorf("must be an origin only, without a path")
	}

	if parsed.RawQuery != "" {
		return fmt.Errorf("must not include a query string")
	}

	if parsed.Fragment != "" {
		return fmt.Errorf("must not include a fragment")
	}

	return nil
}

func validateCookieName(name string) error {
	if strings.ContainsAny(name, " \t\r\n;") {
		return fmt.Errorf("must not contain whitespace or ';'")
	}

	return nil
}

func validateCookieDomain(domain string) error {
	if strings.Contains(domain, "://") {
		return fmt.Errorf("must be a domain only, without a scheme")
	}

	if strings.Contains(domain, "/") {
		return fmt.Errorf("must not contain a path")
	}

	if strings.ContainsAny(domain, " \t\r\n") {
		return fmt.Errorf("must not contain whitespace")
	}

	return nil
}

func validateS3Endpoint(raw string) error {
	raw = strings.TrimSpace(raw)

	if raw == "" {
		return fmt.Errorf("is required")
	}

	if strings.Contains(raw, "://") {
		return fmt.Errorf("must not include a scheme; use host or host:port")
	}

	if strings.Contains(raw, "/") {
		return fmt.Errorf("must not include a path")
	}

	if strings.Contains(raw, "?") || strings.Contains(raw, "#") {
		return fmt.Errorf("must not include query or fragment")
	}

	parsed, err := url.Parse("//" + raw)
	if err != nil {
		return fmt.Errorf("is invalid: %w", err)
	}

	if parsed.Host == "" || parsed.Hostname() == "" {
		return fmt.Errorf("must include a host")
	}

	return nil
}
