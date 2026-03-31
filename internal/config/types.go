package config

import "time"

// Config groups all runtime configuration sections used by the application.
type Config struct {
	App     AppConfig
	HTTP    HTTPConfig
	DB      DBConfig
	S3      S3Config
	SMTP    SMTPConfig
	OAuth   OAuthConfig
	Redis   RedisConfig
	SIEM    SIEMConfig
	JWT     JWTConfig
	Cookie  CookieConfig
	Captcha CaptchaConfig
}

// AppConfig contains generic application metadata and runtime environment settings.
type AppConfig struct {
	// Name is the application name written to logs and diagnostics.
	Name string

	// Env is the current runtime environment, for example dev, test, or prod.
	Env string
}

// HTTPConfig contains settings for the HTTP server and public application URL.
type HTTPConfig struct {
	// APIAddr is the listen address used by the HTTP server, for example ":8080".
	APIAddr string

	// BaseURL is the public base URL used when generating absolute links.
	BaseURL string

	// FrontendOrigin is the allowed CORS origin.
	FrontendOrigin string
}

// DBConfig contains database connection settings.
type DBConfig struct {
	// URL is the PostgreSQL connection string used by the application.
	URL string
}

// S3Config contains object storage settings used by local MinIO and production R2.
type S3Config struct {
	// Endpoint is the S3-compatible API endpoint.
	Endpoint string

	// Bucket is the bucket name used to store application files.
	Bucket string

	// AccessKeyID is the S3 access key identifier.
	AccessKeyID string

	// SecretAccessKey is the S3 secret access key.
	SecretAccessKey string

	// Region is the storage region name expected by the S3-compatible client.
	Region string

	// UsePathStyle enables path-style bucket addressing instead of virtual-host style.
	UsePathStyle bool

	UseSSL bool
}

// SMTPConfig contains email delivery settings.
type SMTPConfig struct {
	// Host is the SMTP server hostname.
	Host string

	// Port is the SMTP server port.
	Port int

	// User is the SMTP username used for authentication.
	User string

	// Password is the SMTP password used for authentication.
	Password string

	// From is the default sender address used by outgoing emails.
	From string
}

// OAuthConfig contains third-party OAuth credentials.
type OAuthConfig struct {
	// GoogleClientID is the Google OAuth client identifier.
	GoogleClientID string

	// GoogleClientSecret is the Google OAuth client secret.
	GoogleClientSecret string

	// GoogleRedirectURL is a callback URI for OAuth.
	GoogleRedirectURL string
}

// CaptchaConfig contains CAPTCHA integration settings.
type CaptchaConfig struct {
	// RecaptchaSiteKey is the public site key used by the frontend.
	RecaptchaSiteKey string

	// RecaptchaSecretKey is the private server-side secret used for verification.
	RecaptchaSecretKey    string
	FailedCodeAttemptsTTL time.Duration
	CodeCaptchaThreshold  int
}

// RedisConfig contains Redis connection settings.
type RedisConfig struct {
	// URL is the Redis connection string used by the application.
	URL string
}

// SIEMConfig contains logging settings used by the Mini-SIEM integration.
type SIEMConfig struct {
	// LogDir is the directory where append-only JSONL application logs are written.
	LogDir string
}

// JWTConfig contains settings used to sign and validate authentication tokens.
type JWTConfig struct {
	// Secret is the signing key used to create and verify JWT tokens.
	Secret string

	// Issuer identifies the backend that issued the token.
	Issuer string

	// Audience identifies the intended recipient or consumer of the token.
	Audience string

	// TTLHours defines how long an issued token remains valid, in hours.
	TTLHours int
}

// CookieConfig contains settings used when writing the authentication cookie.
type CookieConfig struct {
	// Name is the cookie name sent to the client browser.
	Name string

	// Domain is an optional cookie domain attribute.
	// When empty, the cookie is scoped to the current host only.
	Domain string

	// Secure controls whether the cookie is sent only over HTTPS.
	Secure bool
}
