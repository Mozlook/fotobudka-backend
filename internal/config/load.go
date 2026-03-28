package config

import "time"

// Load reads configuration from environment variables, applies defaults,
// and returns a populated Config value.
func Load() (Config, error) {
	smtpPort, err := getEnvInt("SMTP_PORT", 0)
	if err != nil {
		return Config{}, err
	}

	s3UsePathStyle, err := getEnvBool("S3_USE_PATH_STYLE", false)
	if err != nil {
		return Config{}, err
	}

	jwtTTLHours, err := getEnvInt("JWT_TTL_HOURS", 720)
	if err != nil {
		return Config{}, err
	}

	cookieSecure, err := getEnvBool("COOKIE_SECURE", false)
	if err != nil {
		return Config{}, err
	}

	failedCodeAttemptsTTL, err := getEnvDuration("CODE_LOGIN_ATTEMPTS_TTL", time.Minute*10)
	if err != nil {
		return Config{}, err
	}

	codeCaptchaThreshold, err := getEnvInt("CODE_LOGIN_CAPTCHA_THRESHOLD", 2)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "FotoBudka"),
			Env:  getEnv("APP_ENV", "dev"),
		},
		HTTP: HTTPConfig{
			APIAddr:        getEnv("API_ADDR", ":8080"),
			BaseURL:        getEnv("BASE_URL", ""),
			FrontendOrigin: getEnv("FRONTEND_ORIGIN", ""),
		},
		DB: DBConfig{
			URL: getEnv("DB_URL", ""),
		},
		S3: S3Config{
			Endpoint:        getEnv("S3_ENDPOINT", ""),
			Bucket:          getEnv("S3_BUCKET", ""),
			AccessKeyID:     getEnv("S3_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("S3_SECRET_ACCESS_KEY", ""),
			Region:          getEnv("S3_REGION", "auto"),
			UsePathStyle:    s3UsePathStyle,
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", ""),
			Port:     smtpPort,
			User:     getEnv("SMTP_USER", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", ""),
		},
		OAuth: OAuthConfig{
			GoogleClientID:     getEnv("GOOGLE_OAUTH_CLIENT_ID", ""),
			GoogleClientSecret: getEnv("GOOGLE_OAUTH_CLIENT_SECRET", ""),
			GoogleRedirectURL:  getEnv("GOOGLE_OAUTH_REDIRECT_URL", ""),
		},
		Captcha: CaptchaConfig{
			RecaptchaSiteKey:      getEnv("RECAPTCHA_SITE_KEY", ""),
			RecaptchaSecretKey:    getEnv("RECAPTCHA_SECRET_KEY", ""),
			FailedCodeAttemptsTTL: failedCodeAttemptsTTL,
			CodeCaptchaThreshold:  codeCaptchaThreshold,
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", ""),
		},
		SIEM: SIEMConfig{
			LogDir: getEnv("SIEM_LOG_DIR", ""),
		},
		JWT: JWTConfig{
			Secret:   getEnv("JWT_SECRET", ""),
			Issuer:   getEnv("JWT_ISSUER", ""),
			Audience: getEnv("JWT_AUDIENCE", ""),
			TTLHours: jwtTTLHours,
		}, Cookie: CookieConfig{
			Name:   getEnv("COOKIE_NAME", "fotobudka_session"),
			Domain: getEnv("COOKIE_DOMAIN", ""),
			Secure: cookieSecure,
		},
	}

	return cfg, nil
}
