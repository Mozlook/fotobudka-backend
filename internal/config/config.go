package config

import "os"

type Config struct {
	AppName string
	AppEnv  string
	APIAddr string
}

func Load() (Config, error) {
	cfg := Config{
		AppName: getEnv("APP_NAME", "FotoBudka"),
		AppEnv:  getEnv("APP_ENV", "dev"),
		APIAddr: getEnv("API_ADDR", ":8080"),
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
