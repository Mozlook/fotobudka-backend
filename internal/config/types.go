package config

type Config struct {
	App     AppConfig
	HTTP    HTTPConfig
	DB      DBConfig
	S3      S3Config
	SMTP    SMTPConfig
	OAuth   OAuthConfig
	Captcha CaptchaConfig
	Redis   RedisConfig
	SIEM    SIEMConfig
}

type AppConfig struct {
	Name string
	Env  string
}

type HTTPConfig struct {
	APIAddr string
	BaseURL string
}

type DBConfig struct {
	URL string
}

type S3Config struct {
	Endpoint        string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	UsePathStyle    bool
}

type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

type OAuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
}

type CaptchaConfig struct {
	RecaptchaSiteKey   string
	RecaptchaSecretKey string
}

type RedisConfig struct {
	URL string
}

type SIEMConfig struct {
	LogDir string
}
